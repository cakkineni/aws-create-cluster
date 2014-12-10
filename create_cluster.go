package main

//brew install bazaar
//go get launchpad.net/goamz/ec2

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
	"os"
	"strconv"
	//"time"
	"fmt"
	"time"
)

var (
	location,
	keyName,
	cloudConfigCluster,
	cloudConfigAgent,
	amiName,
	size string
	serverCount int
	amzClient   *ec2.EC2
)

func main() {

	login()

	cloudConfigCluster = createCloudConfigCluster()

	privateKey, publicKey := createSshKey()
	cloudConfigAgent = createCloudConfigAgent(publicKey)

	//create coreos servers
	var coreOSCluster ec2.Instance
	for i := 0; i < serverCount; i++ {
		coreOSCluster = createCoreOSServer(i + 1).Instances[0]
	}

	//create agent server
	var pmxAgent ec2.Instance
	pmxAgent = createAgentServer().Instances[0]

	for {
		println("Agent State:" + pmxAgent.State.Name + ", Cluster Node State: " + coreOSCluster.State.Name)
		if pmxAgent.State.Code == 16 && coreOSCluster.State.Code == 16 {
			break
		}
		time.Sleep(10 * time.Second)
		resp, err := amzClient.Instances([]string{pmxAgent.InstanceId, coreOSCluster.InstanceId}, &ec2.Filter{})
		if err != nil {
			panic(err)
		}
		pmxAgent = resp.Reservations[0].Instances[0]
		coreOSCluster = resp.Reservations[1].Instances[0]
	}

	agentIp := pmxAgent.IPAddress
	fleetIP := coreOSCluster.PrivateIPAddress

	setEtcdKey("agent-pri-ssh-key", base64.StdEncoding.EncodeToString([]byte(privateKey)))
	setEtcdKey("agent-fleet-api", fleetIP)
	setEtcdKey("agent-public-ip", agentIp)

	fmt.Scanln()
	time.Sleep(2000 * time.Hour)

}

func init() {
	serverCount, _ = strconv.Atoi(os.Getenv("NODE_COUNT"))
	apiToken := os.Getenv("AWS_ACCESS_KEY_ID")
	apiPassword := os.Getenv("AWS_SECRET_ACCESS_KEY")
	location = os.Getenv("REGION")
	keyName = os.Getenv("SSH_KEY_NAME")
	size = os.Getenv("VM_SIZE")

	var amis []struct {
		Region string
		AMI    string
	}

	amiFile, _ := ioutil.ReadFile("aws_ami.json")
	json.Unmarshal(amiFile, &amis)

	for _, ami := range amis {
		if ami.Region == location {
			amiName = ami.AMI
			break
		}
	}

	println("AMI Used:" + amiName)

	if apiToken == "" || apiPassword == "" || serverCount == 0 || location == "" || size == "" || amiName == "" {
		panic("\n\nMissing Params Or No Matching AMI found...Check Docs...\n\n")
	}
}

func login() {
	println("\nLogging in....")
	auth, err := aws.EnvAuth()

	if err != nil {
		panic(err)
	}

	amzClient = ec2.New(auth, aws.USEast)

}

func createCoreOSServer(id int) *ec2.RunInstancesResp {
	println("Create CoreOS Server")
	createReq := &ec2.RunInstances{
		ImageId:      amiName,
		InstanceType: size,
		UserData:     []byte(cloudConfigAgent),
		MinCount:     serverCount,
		MaxCount:     serverCount,
	}

	return createServer(createReq)
}

func createAgentServer() *ec2.RunInstancesResp {
	println("Create CoreOS Agent Server")
	createReq := ec2.RunInstances{
		ImageId:      amiName,
		InstanceType: size,
		UserData:     []byte(cloudConfigAgent),
		MinCount:     1,
		MaxCount:     1,
	}

	return createServer(&createReq)
}

func createServer(createRequest *ec2.RunInstances) *ec2.RunInstancesResp {
	var resp *ec2.RunInstancesResp
	resp, err := amzClient.RunInstances(createRequest)

	if err != nil {
		panic(err.Error())
	}

	return resp
}
