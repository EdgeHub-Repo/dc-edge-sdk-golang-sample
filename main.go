package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	SDK "github.com/EdgeHub-Repo/dc-edge-sdk-golang"
)

func main() {

	go func() {
		for {
			runtime.NumGoroutine()
			// fmt.Printf("goroutine num = %d\n", NumGoroutine)
			time.Sleep(5 * time.Second)
		}
	}()

	quit := make(chan bool)

	options := SDK.NewEdgeAgentOptions()

	options.DataRecover = true
	options.NodeID = "98e1c5d0-6c19-11ef-817e-750e252956f7"
	options.ConnectType = SDK.ConnectType["DCCS"]
	options.DCCS.Key = "4153cc4ee81168e01d4f83bca76ceeia"
	options.DCCS.URL = "http://api-dccs-ensaas.isghpc.wise-paas.com/"

	interval := 1
	var timer chan bool = nil

	agent := SDK.NewAgent(options)
	agent.SetOnConnectHandler(func(a SDK.Agent) {
		fmt.Println("connect successfully")

		config := generateConfig()
		action := SDK.Action["Create"]
		_ = agent.UploadConfig(action, config)

		status := generateDeviceStatus()
		_ = agent.SendDeviceStatus(status)

		if timer == nil {
			timer = setInterval(func() {
				data := generateData()
				ok := agent.SendData(data)
				if ok {
					fmt.Println(data)
				}
			}, interval, true)
		}
	})
	agent.SetOnDisconnectHandler(func(a SDK.Agent) {
		fmt.Println("disconnect successfully")
	})
	agent.SetOnMessageReceiveHandler(func(args SDK.MessageReceivedEventArgs, a SDK.Agent) {
		msgType := args.Type
		message := args.Message
		switch msgType {
		case SDK.MessageType["WriteValue"]: // message format: WriteDataMessage
			for _, device := range message.(SDK.WriteDataMessage).DeviceList {
				fmt.Println("DeviceId: ", device.ID)
				for _, tag := range device.TagList {
					fmt.Println("TagName: ", tag.Name, ", Value: ", tag.Value)
				}
			}
		case SDK.MessageType["ConfigAck"]: // message format: ConfigAckMessage
			fmt.Println(message.(SDK.ConfigAckMessage).Result)
		case SDK.MessageType["TimeSync"]: //message format: TimeSyncMessage
			fmt.Println(message.(SDK.TimeSyncMessage).UTCTime)
		}
	})

	err := agent.Connect()
	if err != nil {
		fmt.Println(err)
	}
	<-quit
}

func generateConfig() SDK.EdgeConfig {
	nodeConfig := generateNodeConfig()
	edgeConfig := SDK.EdgeConfig{
		Node: nodeConfig,
	}
	return edgeConfig
}

func generateNodeConfig() SDK.NodeConfig {
	var deviceNum = 1

	nodeConfig := SDK.NewNodeConfig()
	nodeConfig.SetType(SDK.EdgeType["Gateway"])

	for idx := 0; idx < deviceNum; idx++ {
		config := generateDeviceConfig(idx + 1)
		nodeConfig.DeviceList = append(nodeConfig.DeviceList, config)
	}
	return nodeConfig
}

func generateDeviceConfig(idx int) SDK.DeviceConfig {
	var deviceID = fmt.Sprintf("%s%d", "Device", idx)
	var analogNum = 3
	var discreteNum = 2
	var textNum = 1

	deviceConfig := SDK.NewDeviceConfig(deviceID)
	deviceConfig.SetName(fmt.Sprintf("%s%d", "Device", idx))
	deviceConfig.SetType("Smart Device")
	deviceConfig.SetDescription(fmt.Sprintf("%s %d", "Device ", idx))

	for idx := 0; idx < analogNum; idx++ {
		config := generateAnalogConfig(idx + 1)
		deviceConfig.AnalogTagList = append(deviceConfig.AnalogTagList, config)
	}
	for idx := 0; idx < discreteNum; idx++ {
		config := generateDiscreteConfig(idx + 1)
		deviceConfig.DiscreteTagList = append(deviceConfig.DiscreteTagList, config)
	}
	for idx := 0; idx < textNum; idx++ {
		config := generateTextConfig(idx + 1)
		deviceConfig.TextTagList = append(deviceConfig.TextTagList, config)
	}
	return deviceConfig
}

func generateAnalogConfig(idx int) SDK.AnalogTagConfig {
	var tagName = fmt.Sprintf("%s%d", "ATag", idx)

	config := SDK.NewAnaglogTagConfig(tagName)
	config.SetDescription(fmt.Sprintf("%s %d", "ATag", idx))
	config.SetReadOnly(false)
	config.SetArraySize(0)
	config.SetSpanHigh(1000.0)
	config.SetSpanLow(0.0)
	config.SetEngineerUnit("")
	config.SetIntegerDisplayFormat(4)
	config.SetFractionDisplayFormat(2)

	return config
}

func generateDiscreteConfig(idx int) SDK.DiscreteTagConfig {
	var tagName = fmt.Sprintf("%s%d", "DTag", idx)

	config := SDK.NewDiscreteTagConfig(tagName)
	config.SetDescription(fmt.Sprintf("%s %d", "DTag ", idx))
	config.SetArraySize(0)
	config.SetReadOnly(true)
	config.SetState0("No")
	config.SetState1("Yes")

	return config
}

func generateTextConfig(idx int) SDK.TextTagConfig {
	var tagName = fmt.Sprintf("%s%d", "TTag", idx)

	config := SDK.NewTextTagConfig(tagName)
	config.SetDescription(fmt.Sprintf("%s %d", "TTag", idx))
	config.SetReadOnly(false)
	config.SetArraySize(0)

	return config
}

func generateData() SDK.EdgeData {
	deviceNum := 1
	msg := SDK.EdgeData{
		Timestamp: time.Now(),
		//Timestamp: time.Date(2020, time.Month(10), 14, 16, 50, 33, 983, time.Local)	// customized time stamp format
	}

	for idx := 0; idx < deviceNum; idx++ {
		analogNum := 3
		discreteNum := 2
		textNum := 1
		deviceID := fmt.Sprintf("%s%d", "Device", idx+1)
		for num := 0; num < analogNum; num++ {
			t := SDK.EdgeTag{
				DeviceID: deviceID,
				TagName:  fmt.Sprintf("%s%d", "ATag", num+1),
				Value:    rand.Float64(),
			}
			//fmt.Println(rand.Float64())

			msg.TagList = append(msg.TagList, t)
		}
		for num := 0; num < discreteNum; num++ {
			t := SDK.EdgeTag{
				DeviceID: deviceID,
				TagName:  fmt.Sprintf("%s%d", "DTag", num+1),
				Value:    rand.Intn(7),
			}
			msg.TagList = append(msg.TagList, t)
		}
		for num := 0; num < textNum; num++ {
			t := SDK.EdgeTag{
				DeviceID: deviceID,
				TagName:  fmt.Sprintf("%s%d", "TTag", num+1),
				Value:    fmt.Sprintf("%s%d", "Test", num+1),
			}
			msg.TagList = append(msg.TagList, t)
		}
	}
	return msg
}

func setInterval(someFunc func(), seconds int, async bool) chan bool {
	interval := time.Duration(seconds) * time.Second
	ticker := time.NewTicker(interval)
	clear := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				if async {
					go someFunc()
				} else {
					someFunc()
				}
			case <-clear:
				ticker.Stop()
				close(clear)
				return
			}
		}
	}()
	return clear
}

func generateDeviceStatus() SDK.EdgeDeviceStatus {
	status := SDK.EdgeDeviceStatus{
		Timestamp: time.Now(),
	}
	deviceNum := 1

	for idx := 0; idx < deviceNum; idx++ {
		s := SDK.DeviceStatus{
			ID:     fmt.Sprintf("%s%d", "Device", idx+1),
			Status: SDK.Status["Online"],
		}
		status.DeviceList = append(status.DeviceList, s)
	}
	return status
}
