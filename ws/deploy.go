package ws

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"graces/exterr"
	"graces/model"
	"graces/rpc"
	"graces/syncer"
	"graces/web/dao"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

const defaultKeyfile = "./keystore"

var (
	DefaultDeploy              *deploy
	DefaultContractInterpreter = "wasm"
	DefaultCover               = " --cover"
	DefaultNoCover             = ""
	DefaultDir                 = "deploy/release/linux/scripts/"
)

func init() {
	DefaultDeploy = newDeploy()
}

func newDeploy() *deploy {
	return &deploy{}
}

type deploy struct {
}

func (d *deploy) DeployContract(chainID string, account string, files []*multipart.FileHeader) (interface{}, error) {
	logrus.Debugf("Contract deployment start.")
	//defer logrus.Debugf("Contract deployment finished.")
	// 1、通过 chainID 获取链信息
	id, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return "", err
	}

	filter := bson.M{
		"_id": id,
	}
	chain, err := dao.DefaultChainDao.Chain(filter)
	if err != nil {
		return "", err
	}

	// 2、通过链信息获取其对应的链 rpc 连接客户端
	cli, err := d.getRPCClientByChain(*chain)
	if err != nil {
		return "", err
	}

	// 3、通过客户端获取消息调用器
	caller := rpc.NewMsgCaller(cli)

	// 4、处理合约文件并创建部署请求信息结构体
	var DeployRequestInfo model.DeployRequest

	var fileBytes = make([][]byte, 2)
	var funcParams = new(model.DeployInfo) //合约bytes

	for i, file := range files {
		dst := path.Join("/tmp", file.Filename)
		//ctx.SaveUploadedFile(file, dst)
		f, _ := file.Open()
		fileBytes[i], _ = ioutil.ReadAll(f)
		if strings.HasSuffix(file.Filename, "json") {
			DeployRequestInfo.AbiPath = dst
			funcParams.AbiBytes = string(fileBytes[i])
		} else if strings.HasSuffix(file.Filename, "wasm") || strings.HasSuffix(file.Filename, "evm") {
			DeployRequestInfo.CodePath = dst
			funcParams.CodeBytes = string(fileBytes[i])
		}

		if strings.HasSuffix(DeployRequestInfo.CodePath, "wasm") {
			DeployRequestInfo.Interpreter = "wasm"
		} else if strings.HasSuffix(DeployRequestInfo.CodePath, "evm") {
			DeployRequestInfo.Interpreter = "evm"
		}
	}

	if DeployRequestInfo.CodePath == "" || DeployRequestInfo.AbiPath == "" || DeployRequestInfo.Interpreter == "" {
		msg := fmt.Sprintf("something missing!")
		err := exterr.NewError(exterr.ErrCodeParameterInvalid, msg)
		return "", err
	}

	// 5. 创建交易结构体
	txParams := &rpc.TxParams{}
	//fi, err := os.Open("deploy/release/deployment_conf/" + chain.Name + "/global/keyfile.account")
	//if err != nil {
	//	logrus.Errorln(err)
	//} else {
	//	br := bufio.NewReader(fi)
	//	for {
	//		a, _, c := br.ReadLine()
	//		if c == io.EOF {
	//			break
	//		}
	//		txParams.From = string(a)
	//	}
	//}
	//defer fi.Close()
	txParams.From = account

	//6. 生合约参数对象
	contractParams := d.buildDeployContractsParams(DeployRequestInfo.Interpreter, funcParams)
	deployedContract, err := caller.DeployContract(txParams, contractParams)
	if err != nil {
		logrus.Debug(err)
	}

	//logrus.Info(deployedContract)

	return deployedContract, nil
}

func (d *deploy) buildDeployContractsParams(interpreter string, contractParams *model.DeployInfo) *rpc.ContractParams {
	if interpreter == "" {
		interpreter = DefaultContractInterpreter
	}

	// 构造 getRegisteredContracts() 函数的查询条件参数
	funcParams := &struct {
		CodeBytes string
		AbiBytes  string
	}{}
	funcParams.CodeBytes = contractParams.CodeBytes
	funcParams.AbiBytes = contractParams.AbiBytes

	// 构造合约参数
	contract := &rpc.ContractParams{
		ContractAddr: "",
		Method:       "",
		AbiMethods:   nil,
		Interpreter:  interpreter,
		Data:         funcParams,
	}
	return contract
}

func (d *deploy) getRPCClientByChain(chain model.Chain) (*rpc.Client, error) {
	host := fmt.Sprintf("%v:%v", chain.IP, chain.RPCPort)
	uri := url.URL{
		Scheme: "http",
		Host:   host,
	}

	nodeConfig, ok := chain.ChainConfig["node"].(map[string]interface{})
	if !ok {
		return nil, errors.New("chain [%v] without node config")
	}

	keyfilePath := nodeConfig["keyfile_path"].(string)
	passphrase := nodeConfig["passphrase"].(string)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := rpc.NewClient(ctx, uri.String(), passphrase, keyfilePath)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (d *deploy) DeployNewNode(chainInfo interface{}, c chan []byte) {
	//projectName := chainInfo.(map[string]interface{})["projectName"].(string)
	//filter := bson.M{"name": projectName}
	//chain, err := dao.DefaultChainDao.Chain(filter)
	//if err != nil {
	//	return
	//}
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when deploying node. Error: %s", err)
		}
		//todo 往channel里面写东西（成功时
	}()

	chainIDStr := chainInfo.(map[string]interface{})["chainID"].(string)
	chainID, err := primitive.ObjectIDFromHex(chainIDStr)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	filter := bson.M{"_id": chainID}
	chain, err := dao.DefaultChainDao.Chain(filter)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	filter = bson.M{}
	if err != nil {
		return
	}
	filter["chain_id"] = chainID
	findOps := options.Count()
	nodeCount, err := dao.DefaultNodeDao.Count(filter, findOps)
	if err != nil {
		logrus.Error(err)
		return
	}

	remoteIP := chainInfo.(map[string]interface{})["remoteIP"].(string)
	userName := chainInfo.(map[string]interface{})["remoteName"].(string)
	deployCount := 1 // by default
	deployCount, _ = strconv.Atoi(chainInfo.(map[string]interface{})["count"].(string))
	//nodeID := chainInfo.(map[string]interface{})["nodeID"].(string)

	for i := 0; i < deployCount; i++ {
		//todo 改用context来判读prepare中，go协程是否结束，或者waitgroup?
		d.Prepare(chain.Name, remoteIP, userName, DefaultDir, DefaultNoCover, strconv.FormatInt(nodeCount+int64(i), 10), c)
		//time.Sleep(time.Duration(2) * time.Second)
		//d.Start(*chain, DefaultDir, nodeID, c)
		d.DeployNode(*chain, DefaultDir, strconv.FormatInt(nodeCount+int64(i), 10), c)
	}

	//todo 返回部署成功的节点数量
	return
}

func (d *deploy) DeployNode(chainInfo model.Chain, dir string, nodeID string, c chan []byte) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when deploying node for chain %s. Error: %s", chainInfo.Name, err)
		}
		//todo 往channel里面写东西（成功时
	}()

	c <- []byte("Start new node for " + chainInfo.Name + "!")
	//todo 根据chaininfo.name来查找对应的链信息，然后创建节点model，作链id与节点id的对应。然后insertNode, 然后return（不用做后续操作，因为链信息已入库。
	filter := bson.M{"name": chainInfo.Name}
	chain, err := dao.DefaultChainDao.Chain(filter)
	if err != nil {
		return
	}

	chainVO, err := chain.ToVO()
	if err != nil {
		return
	}
	//todo optimize it
	cmd := "./deploy.sh -p " + chainInfo.Name + " -n " + nodeID
	cmdDeployNewNode := exec.Command("sh", "-c", cmd)
	cmdDeployNewNode.Dir = dir
	d.ShellCall(cmdDeployNewNode, nil, c)

	syncInfo := syncer.DefaultChainDataSyncManager.BuildChainSyncInfo(chainVO.ID)
	if syncInfo.NodeDataSyncInfo != nil && syncInfo.NodeDataSyncInfo.Status == syncer.StatusSyncing {
		logrus.Infof("this chain[%s] node data is syncing, don't repeat sync for it", chainVO.ID)
		return
	}

	err = syncer.DefaultChainDataSyncManager.SyncNode(chainVO.ID, false)
	if err != nil {
		syncInfo.NodeDataSyncInfo.ErrMsg = err.Error()
		syncInfo.NodeDataSyncInfo.Status = syncer.StatusError
		syncer.DefaultChainDataSyncManager.ErrChan <- &model.SyncErrMsg{
			ChainID: chainVO.ID,
			ErrType: syncer.ErrTypeBlockOrTXSync,
			Err:     err,
		}

		return
	}

	logrus.Debugf("chain[%v] data sync success", chainVO.ID)
	return
}

func (d *deploy) nameByChainID(chainID string) (string, error) {
	objectId, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return "", exterr.ErrObjectIDInvalid
	}
	filter := bson.M{"_id": objectId}
	chain, err := dao.DefaultChainDao.Chain(filter)
	if err != nil {
		return "", exterr.NewError(exterr.ErrCodeFind, err.Error())
	}

	return chain.Name, nil
}

func (d *deploy) StopNode(info interface{}, c chan []byte) error {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when stoping node. Error: %s", err)
		}
		//todo 往channel里面写东西（成功时
	}()

	//projectName := info.(map[string]interface{})["projectName"].(string)
	chainID := info.(map[string]interface{})["chainID"].(string)
	nodeID := info.(map[string]interface{})["nodeID"].(string)

	chainName, err := d.nameByChainID(chainID)
	if err != nil {
		return err
	}

	//todo 判断该id对应的节点是否已经部署，若没部署，则返回err

	cmd := "./clear.sh -p " + chainName + " -n " + nodeID + " -m stop"
	cmdStop := exec.Command("sh", "-c", cmd)
	cmdStop.Dir = DefaultDir

	go d.ShellCall(cmdStop, nil, c)

	return nil
}

func (d *deploy) StartNode(info interface{}, c chan []byte) error {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when starting node. Error: %s", err)
		}
		//todo 往channel里面写东西（成功时
	}()

	//projectName := info.(map[string]interface{})["projectName"].(string)
	chainID := info.(map[string]interface{})["chainID"].(string)
	nodeID := info.(map[string]interface{})["nodeID"].(string)

	chainName, err := d.nameByChainID(chainID)
	if err != nil {
		return err
	}

	cmd := "./start.sh -p " + chainName + " -n " + nodeID
	cmdStart := exec.Command("sh", "-c", cmd)
	cmdStart.Dir = DefaultDir

	go d.ShellCall(cmdStart, nil, c)

	return nil
}

func (d *deploy) RestartNode(info interface{}, c chan []byte) error {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when restarting node. Error: %s", err)
		}
		//todo 往channel里面写东西（成功时
	}()

	//projectName := info.(map[string]interface{})["projectName"].(string)
	chainID := info.(map[string]interface{})["chainID"].(string)
	nodeID := info.(map[string]interface{})["nodeID"].(string)

	chainName, err := d.nameByChainID(chainID)
	if err != nil {
		return err
	}

	cmdc := "./clear.sh -p " + chainName + " -n " + nodeID + " -m stop"
	cmdStop := exec.Command("sh", "-c", cmdc)
	cmdStop.Dir = DefaultDir

	cmds := "./start.sh -p " + chainName + " -n " + nodeID
	cmdStart := exec.Command("sh", "-c", cmds)
	cmdStart.Dir = DefaultDir

	go func() {
		d.ShellCall(cmdStop, nil, c)
		d.ShellCall(cmdStart, nil, c)
	}()

	return nil
}

func (d *deploy) DeployNewChain(chainInfo interface{}, c chan []byte) {
	projectName := chainInfo.(map[string]interface{})["projectName"].(string)
	remoteIP := chainInfo.(map[string]interface{})["remoteIP"].(string)
	userName := chainInfo.(map[string]interface{})["remoteName"].(string)

	//异步部署新链
	go d.Prepare(projectName, remoteIP, userName, DefaultDir, DefaultCover, "", c)

	//todo 返回调用成功信息给前端？
	return
}

//todo prepare, transfer, init三个函数相同代码块比较多，待优化。
func (d *deploy) Prepare(projectName string, remoteIp string, userName string, dir string, cover string, nodeID string, c chan []byte) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when preparing files for node. Error: %s", err)
		}
		//todo 往channel里面写东西（成功时
	}()

	cmd := "./prepare.sh -p " + projectName + " -a " + userName + "@" + remoteIp
	if cover == DefaultCover {
		cmd += cover
	}
	cmdPrepare := exec.Command("sh", "-c", cmd)
	cmdPrepare.Dir = dir

	if nodeID != "" { // nodeID != "" 则为现有链创建新节点
		stdin, err := cmdPrepare.StdinPipe()
		if err != nil {
			c <- []byte("Failed to create stdin pipe.")
			log.Fatalf("failed to create stdin pipe: %v", err)
			return
		}
		c <- []byte("Prepare create new node for " + projectName)
		d.ShellCall(cmdPrepare, stdin, c) //no cover -- enter "n" to stdin for "yesOrNo"
	} else {
		d.ShellCall(cmdPrepare, nil, c)
		c <- []byte("Prepare files for " + projectName + " success!")

		d.Transfer(model.Chain{
			Name:     projectName,
			IP:       remoteIp,
			Username: userName,
		}, dir, cover, c)
	}
	//ws.Socket.WriteMessage(websocket.BinaryMessage, []byte("Prepare success!"))

	return
}

func (d *deploy) Transfer(chaininfo model.Chain, dir string, cover string, c chan []byte) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when transfering files for chain %s. Error: %s", chaininfo.Name, err)
		}
		//todo 往channel里面写东西（成功时
	}()

	cmd := "./transfer.sh -p " + chaininfo.Name
	cmdTransfer := exec.Command("sh", "-c", cmd)
	cmdTransfer.Dir = dir

	d.ShellCall(cmdTransfer, nil, c)
	//ws.Socket.WriteMessage(websocket.BinaryMessage, []byte("Transfer success!"))
	if cover == "" {
		c <- []byte("Transfer new node files for " + chaininfo.Name + "!")
	} else {
		c <- []byte("Transfer files for " + chaininfo.Name + "!")
	}

	d.Init(chaininfo, dir, cover, c)

	return
}

func (d *deploy) Init(chaininfo model.Chain, dir string, cover string, c chan []byte) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when initializing chain %s. Error: %s", chaininfo.Name, err)
		}
		//todo 往channel里面写东西（成功时
	}()

	cmd := "./init.sh -p " + chaininfo.Name
	cmdInit := exec.Command("sh", "-c", cmd)
	cmdInit.Dir = dir

	d.ShellCall(cmdInit, nil, c)
	//ws.Socket.WriteMessage(websocket.BinaryMessage, []byte("Init success!"))
	if cover == "" {
		c <- []byte("Initializing new node for " + chaininfo.Name + "!")
	} else {
		c <- []byte("Initializing " + chaininfo.Name + "!")
	}

	d.Start(chaininfo, dir, cover, c)

	return
}

func (d *deploy) Start(chaininfo model.Chain, dir string, cover string, c chan []byte) error {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Something error when starting chain %s. Error: %s", chaininfo.Name, err)
		}
		//todo 往channel里面写东西（成功时
	}()

	cmd := "./start.sh -p " + chaininfo.Name
	cmdStart := exec.Command("sh", "-c", cmd)
	cmdStart.Dir = dir

	d.ShellCall(cmdStart, nil, c)
	//c <- []byte("Start " + chaininfo.Name + " success!")

	fi, err := os.Open("deploy/release/deployment_conf/" + chaininfo.Name + "/deploy_node-0.conf")
	logrus.Info("deploy/release/deployment_conf/" + chaininfo.Name + "/deploy_node-0.conf")
	if err != nil {
		logrus.Errorln(err)
	} else {
		br := bufio.NewReader(fi)
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			if strings.Contains(string(a), "rpc_port") {
				rpcPortInt, _ := strconv.Atoi(string(a)[len(a)-4:])
				chaininfo.RPCPort = uint64(rpcPortInt)
				P2PPortTemp, _ := strconv.Atoi("1" + strconv.FormatUint(chaininfo.RPCPort, 10))
				chaininfo.P2PPort = uint64(P2PPortTemp)
				tempWSPort, _ := strconv.Atoi("2" + strconv.FormatUint(chaininfo.RPCPort, 10))
				chaininfo.WSPort = uint64(tempWSPort)
			}
			//if strings.Contains(string(a), "deploy_path") {
			//	chaininfo.Path = string(a)[12:]
			//}
		}
	}
	err = fi.Close()
	if err != nil {
		logrus.Debug(err)
	}

	chaininfo.ID = primitive.NewObjectID()

	if chaininfo.ChainConfig == nil {
		chaininfo.ChainConfig = make(map[string]interface{})
		chaininfo.ChainConfig["ws"] = map[string]interface{}{}
		chaininfo.ChainConfig["node"] = map[string]interface{}{}

		chaininfo.ChainConfig["node"].(map[string]interface{})["keyfile_path"] = "./keystore"
		chaininfo.ChainConfig["node"].(map[string]interface{})["passphrase"] = "0"

		wsConfig := make(map[string]interface{})
		wsConfig["port"] = "2" + strconv.FormatUint(chaininfo.RPCPort, 10)
		//wsConfig["path"] = chaininfo.Path

		wsTopicsConfig := make(map[string]interface{})
		newHeadsConfig := make(map[string]interface{})

		newHeadsDetail := make(map[string]interface{})
		newHeadsDetail["name"] = "newHeads"
		newHeadsDetail["params"] = "{\"jsonrpc\":\"2.0\",\"method\":\"eth_subscribe\", \"params\": [\"newHeads\"],\"id\":\"subscription\"}"

		newHeadsConfig["new_heads"] = newHeadsDetail
		wsTopicsConfig["topics"] = newHeadsConfig

		chaininfo.ChainConfig["ws"].(map[string]interface{})["path"] = wsConfig["path"]
		chaininfo.ChainConfig["ws"].(map[string]interface{})["port"] = wsConfig["port"]

		chaininfo.ChainConfig["ws"].(map[string]interface{})["topics"] = wsTopicsConfig["topics"]
	}

	logrus.Debugf("chain %v start completed!", chaininfo.ID)

	//todo 待优化
	p2pPort, _ := strconv.Atoi("1" + strconv.FormatUint(chaininfo.RPCPort, 10))
	primaryNode := model.Node{
		ID:         primitive.NewObjectID(),
		ChainID:    chaininfo.ID,
		Name:       "0",
		ExternalIP: chaininfo.IP,
		RPCPort:    int(chaininfo.RPCPort),
		P2PPort:    p2pPort,
	}
	err = dao.DefaultNodeDao.InsertNode(primaryNode)
	if err != nil {
		c <- []byte("Insert Node err!")
		return err
	}

	err = dao.DefaultChainDao.InsertChain(chaininfo)
	if err != nil {
		c <- []byte("Insert Node err!")
		return err
	}

	err = DefaultWSSubscriber.SubTopicsForChain(&chaininfo)
	if err != nil {
		c <- []byte("Insert Node err!")
		return err
	}

	c <- []byte("Start " + chaininfo.Name + " success!")

	return nil
}

func (d *deploy) ShellCall(cmd *exec.Cmd, stdin io.WriteCloser, c chan []byte) error {
	stdout, _ := cmd.StdoutPipe()
	reader := bufio.NewReader(stdout)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("failed to call Run(): %v", err)
	}

	if stdin != nil {
		_, err := io.WriteString(stdin, "n\n")
		if err != nil {
			logrus.Fatal(err)
			return err
		}
	}

	tempLine := ""

	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")
		c <- []byte(line)

		if err != nil || io.EOF == err {
			logPrint, complete := d.HandleLogs(tempLine)
			if complete {
				logrus.Info(logPrint)
			}
			break
		}
		//logrus.Info(line)
		tempLine = line
	}

	return err
}

func (d *deploy) HandleLogs(log string) (string, bool) {
	if strings.Contains(log, "completed") {
		logSplit := strings.Split(log, ":")
		return strings.TrimSpace(logSplit[1]), true
	}

	return "Not completed, something wrong!", false
}
