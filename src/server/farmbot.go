package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	async "github.com/FarmbotSimulator/farmbotProxy/src/async"
	mqtt_ "github.com/eclipse/paho.mqtt.golang"
)

/**
 * publish response from device
 * @param {object} args - receivedMessage.args
 */
func publishFromDevice(client mqtt_.Client, botId string, args map[string]interface{}) {
	message := map[string]interface{}{
		"args": map[string]interface{}{
			"label": args["label"],
		},
		"kind": "rpc_ok",
	}
	publishGeneralData(client, botId, message, "from_device")
}

/**
 * Process messages published to sync topic
 * @param {string} wholeTopic - mqtt topic
 * @param {string} message - received message
 */
func processSyncMessages(client mqtt_.Client, botId string, wholeTopic string, message map[string]interface{}) {
	syncType := strings.Split(wholeTopic, "/")[3]
	switch syncType {
	case "FarmwareEnv":
		publishFromDevice(client, botId, message["args"].(map[string]interface{}))
		break
	case "Device":
		publishFromDevice(client, botId, message["args"].(map[string]interface{}))
		break
	}
}

/**
 * setUserEnv
 * @param {Array} body
 */
func setUserEnv(botId string, body []interface{}) {
	for _, bodyPart := range body {
		kind := bodyPart.(map[string]interface{})["kind"]
		switch kind {
		case "pair":
			args := bodyPart.(map[string]interface{})["args"].(map[string]interface{})
			userEnv[botId] = make(map[string]interface{})
			userEnv[botId].(map[string]interface{})[args["label"].(string)] = args["value"]
			break
		}
	}
}

/**
 * Process rpc_requests
 * @param {Array} messages - array of rpc actions
 * @param {Object} args - args from message received
 */
func processRpcRequests(client mqtt_.Client, botId string, messages []interface{}, args map[string]interface{}) {
	for _, message := range messages {
		kind := message.(map[string]interface{})["kind"]
		switch kind {
		case "set_user_env":
			setUserEnv(botId, message.(map[string]interface{})["body"].([]interface{}))
			publishFromDevice(client, botId, args)
			break
		case "move_absolute":
			moveAbsoluteRelative(client, botId, message.(map[string]interface{})["args"].(map[string]interface{}), "absolute")
			break
		case "move_relative":
			moveAbsoluteRelative(client, botId, message.(map[string]interface{})["args"].(map[string]interface{}), "relative")
			break
		}
	}
}

func moveAbsoluteRelative(client mqtt_.Client, botId string, position map[string]interface{}, moveType string) {
	locationData := make(map[string]interface{})
	if moveType == "absolute" {
		locationData = position["location"].(map[string]interface{})["args"].(map[string]interface{})
	} else {
		locationData = position
	}

	x := float32(locationData["x"].(float64))
	y := float32(locationData["y"].(float64))
	z := float32(locationData["z"].(float64))
	speed := float32(position["speed"].(float64))

	logDataStr := `{
		"channels": [],
		"created_at": "new Date().getTime().toString().slice(0, 10)",
		"major_version": 10,
		"message": "Farmbot is up and running!",
		"meta": {
			"assertion_passed": null,
			"assertion_type": null
		},
		"minor_version": 1,
		"patch_version": 4,
		"type": "success",
		"verbosity": 1,
		"x": "this.location.x",
		"y": "this.location.y",
		"z": "this.location.z"
	}`
	var logData map[string]interface{}
	json.Unmarshal([]byte(logDataStr), &logData)
	logData["created_at"] = fmt.Sprintf("%v", time.Now().Unix())
	if moveType == "absolute" {
		logData["message"] = fmt.Sprintf("Moving to (%f,%f,%f)", x, y, z)
	} else {
		logData["message"] = fmt.Sprintf("Moving relative to (%f,%f,%f)", x, y, z)
	}
	logData["x"] = botStatus[botId].(map[string]float32)["x"]
	logData["y"] = botStatus[botId].(map[string]float32)["y"]
	logData["z"] = botStatus[botId].(map[string]float32)["z"]
	publishGeneralData(client, botId, logData, "logs")

	tmpData := map[string]float32{
		"speed": speed,
		"x":     x,
		"y":     y,
		"z":     z,
	}
	tmp, _ := json.Marshal(tmpData)
	topic := fmt.Sprintf("/%s/move_relative", botId)
	if moveType == "absolute" {
		topic = fmt.Sprintf("/%s/move_absolute", botId)
	}
	server.Publish(topic, tmp, false)
}

/**
 * Process messages published to from_clients topic
 * @param {string} wholeTopic - mqtt topic
 * @param {string} message - received message
 */
func processFromClientTopicMessages(client mqtt_.Client, botId string, wholeTopic string, message []byte) {
	// } catch (error) {}
	async.Exec(func() interface{} {
		messageStr := string(message)
		var messageInterface map[string]interface{}
		json.Unmarshal([]byte(messageStr), &messageInterface)
		msgType := messageInterface["kind"]
		body := messageInterface["body"].([]interface{})
		args := messageInterface["args"].(map[string]interface{})
		// fmt.Println(msgType)
		// fmt.Println(body)
		// fmt.Println(args)
		switch msgType {
		case "rpc_request":
			processRpcRequests(client, botId, body, args)
			break
		}
		return nil
	})
}

func monitorDownlinkMessages(client mqtt_.Client, wholeTopic string, message []byte) {
	// messageStr := string(message)
	botId := strings.Split(wholeTopic, "/")[1]
	topic := strings.Split(wholeTopic, "/")[2]
	switch topic {
	case "ping":
		sendPingPong(botId, client, string(message))
		// client.Publish(`bot/`+botId+`/pong/`+string(message), 0, false, message) // check. Move to sendPingPong
		break
	case "from_clients":
		processFromClientTopicMessages(client, botId, wholeTopic, message)
		break
	case "sync":
		messageStr := string(message)
		var messageInterface map[string]interface{}
		json.Unmarshal([]byte(messageStr), &messageInterface)
		processSyncMessages(client, botId, wholeTopic, messageInterface)
		break
	}
}

/**
 * Publish general messages
 * @param {string||object} msgData
 * @param {string} msgType
 */
func publishGeneralData(client mqtt_.Client, botId string, msgData map[string]interface{}, msgType string) {
	msgDataBytes, _ := json.Marshal(msgData)
	msgDataStr := string(msgDataBytes)
	client.Publish(`bot/`+botId+`/`+msgType, 0, false, msgDataStr)
}

func schedulePublishLogs(client mqtt_.Client, botId string) {
	logDataStr := `{
		"channels": [],
		"created_at": "new Date().getTime().toString().slice(0, 10)",
		"major_version": 10,
		"message": "Farmbot is up and running!",
		"meta": {
			"assertion_passed": null,
			"assertion_type": null
		},
		"minor_version": 1,
		"patch_version": 4,
		"type": "success",
		"verbosity": 1,
		"x": "this.location.x",
		"y": "this.location.y",
		"z": "this.location.z"
	}`
	var logData map[string]interface{}
	json.Unmarshal([]byte(logDataStr), &logData)
	logData["created_at"] = fmt.Sprintf("%v", time.Now().Unix())
	logData["x"] = botStatus[botId].(map[string]float32)["x"]
	logData["y"] = botStatus[botId].(map[string]float32)["y"]
	logData["z"] = botStatus[botId].(map[string]float32)["z"]
	publishGeneralData(client, botId, logData, "logs")
}

//botStatus
func publishStatusMessage(client mqtt_.Client, botId string) {
	statusDataStr := `{
		"configuration": {
			"arduino_debug_messages": false,
			"auto_sync": false,
			"beta_opt_in": false,
			"disable_factory_reset": false,
			"firmware_debug_log": false,
			"firmware_hardware": "farmduino_k14",
			"firmware_input_log": false,
			"firmware_output_log": false,
			"network_not_found_timer": null,
			"os_auto_update": false,
			"sequence_body_log": false,
			"sequence_complete_log": false,
			"sequence_init_log": false
		},
		"informational_settings": {
			"busy": false,
			"cache_bust": null,
			"commit": "1c5ef14bfa90cbbaff792f3a14c7c0707c73bb08",
			"controller_commit": "1c5ef14bfa90cbbaff792f3a14c7c0707c73bb08",
			"controller_uuid": "29417194-a853-55ef-6de8-91dd9b849b0b",
			"controller_version": "10.1.4",
			"cpu_usage": 3,
			"disk_usage": 0,
			"env": "prod",
			"firmware_commit": "1711db1d9923bc295f81a5eb9952f6b3a10db6a9",
			"firmware_version": "6.4.2.G",
			"idle": true,
			"last_status": null,
			"locked": false,
			"memory_usage": 60,
			"node_name": "farmbot@farmbot-000000004ed75c64.local",
			"private_ip": "192.168.100.30",
			"scheduler_usage": 3,
			"soc_temp": 34,
			"sync_status": "sync_now",
			"target": "rpi3",
			"throttled": "0x0",
			"update_available": true,
			"uptime": "this.uptime",
			"wifi_level": -37,
			"wifi_level_percent": 91
		},
		"jobs": {},
		"location_data": {
			"axis_states": {
				"x": "unknown",
				"y": "unknown",
				"z": "unknown"
			},
			"load": {
				"x": null,
				"y": null,
				"z": null
			},
			"position": {
				"x": "this.location.x",
				"y": "this.location.y",
				"z": "this.location.z"
			},
			"raw_encoders": {
				"x": 0.0,
				"y": 0.0,
				"z": 0.0
			},
			"scaled_encoders": {
				"x": 0.0,
				"y": 0.0,
				"z": 0.0
			}
		},
		"mcu_params": {
			"movement_stall_sensitivity_z": 30.0,
			"movement_stop_at_max_y": 0.0,
			"encoder_missed_steps_max_y": 5.0,
			"movement_keep_active_y": 1.0,
			"movement_steps_acc_dec_y": 300.0,
			"movement_invert_2_endpoints_y": 0.0,
			"movement_keep_active_z": 1.0,
			"movement_max_spd_y": 400.0,
			"pin_guard_5_time_out": 60.0,
			"encoder_scaling_z": 5556.0,
			"pin_guard_4_pin_nr": 0.0,
			"pin_guard_3_time_out": 60.0,
			"movement_steps_acc_dec_x": 300.0,
			"encoder_missed_steps_decay_z": 5.0,
			"movement_home_up_x": 0.0,
			"movement_secondary_motor_invert_x": 1.0,
			"encoder_enabled_y": 1.0,
			"movement_axis_nr_steps_x": 0.0,
			"movement_motor_current_x": 600.0,
			"movement_timeout_x": 120.0,
			"movement_invert_endpoints_x": 0.0,
			"movement_home_spd_y": 50.0,
			"encoder_enabled_z": 1.0,
			"movement_enable_endpoints_z": 0.0,
			"movement_home_at_boot_y": 0.0,
			"movement_axis_nr_steps_z": 0.0,
			"movement_invert_motor_x": 0.0,
			"encoder_invert_z": 0.0,
			"movement_home_spd_z": 50.0,
			"encoder_type_y": 0.0,
			"movement_enable_endpoints_y": 0.0,
			"pin_guard_3_active_state": 1.0,
			"encoder_scaling_y": 5556.0,
			"movement_stop_at_max_x": 0.0,
			"encoder_missed_steps_decay_x": 5.0,
			"movement_timeout_z": 120.0,
			"encoder_scaling_x": 5556.0,
			"movement_keep_active_x": 1.0,
			"movement_min_spd_y": 50.0,
			"movement_max_spd_x": 400.0,
			"movement_stop_at_max_z": 0.0,
			"encoder_missed_steps_max_x": 5.0,
			"pin_guard_1_active_state": 1.0,
			"movement_home_up_z": 1.0,
			"encoder_missed_steps_decay_y": 5.0,
			"pin_guard_1_time_out": 60.0,
			"movement_step_per_mm_x": 5.0,
			"movement_home_at_boot_x": 0.0,
			"movement_invert_2_endpoints_z": 0.0,
			"movement_home_spd_x": 50.0,
			"pin_guard_4_active_state": 1.0,
			"movement_stall_sensitivity_x": 30.0,
			"encoder_type_x": 0.0,
			"movement_min_spd_z": 50.0,
			"pin_guard_3_pin_nr": 0.0,
			"pin_guard_2_pin_nr": 0.0,
			"pin_guard_5_active_state": 1.0,
			"pin_guard_2_active_state": 1.0,
			"movement_motor_current_y": 600.0,
			"movement_home_up_y": 0.0,
			"movement_axis_nr_steps_y": 0.0,
			"movement_stall_sensitivity_y": 30.0,
			"movement_invert_endpoints_z": 0.0,
			"movement_home_at_boot_z": 0.0,
			"movement_microsteps_y": 1.0,
			"pin_guard_1_pin_nr": 0.0,
			"movement_invert_motor_z": 0.0,
			"pin_guard_4_time_out": 60.0,
			"encoder_use_for_pos_x": 0.0,
			"pin_guard_5_pin_nr": 0.0,
			"encoder_invert_x": 0.0,
			"movement_step_per_mm_y": 5.0,
			"movement_invert_2_endpoints_x": 0.0,
			"encoder_use_for_pos_y": 0.0,
			"movement_invert_motor_y": 0.0,
			"movement_microsteps_x": 1.0,
			"param_mov_nr_retry": 3.0,
			"movement_min_spd_x": 50.0,
			"movement_invert_endpoints_y": 0.0,
			"movement_steps_acc_dec_z": 300.0,
			"movement_max_spd_z": 400.0,
			"movement_stop_at_home_z": 0.0,
			"param_e_stop_on_mov_err": 0.0,
			"movement_enable_endpoints_x": 0.0,
			"encoder_enabled_x": 1.0,
			"movement_microsteps_z": 1.0,
			"encoder_missed_steps_max_z": 5.0,
			"encoder_invert_y": 0.0,
			"pin_guard_2_time_out": 60.0,
			"movement_step_per_mm_z": 25.0,
			"encoder_type_z": 0.0,
			"movement_timeout_y": 120.0,
			"movement_secondary_motor_x": 1.0,
			"movement_stop_at_home_x": 0.0,
			"movement_motor_current_z": 600.0,
			"movement_stop_at_home_y": 0.0,
			"encoder_use_for_pos_z": 0.0
		},
		"pins": {},
		"process_info": {
			"farmwares": {}
		},
		"user_env": {
			"LAST_CLIENT_CONNECTED": "2020-08-25T06:28:44.168Z",
			"camera": "\"USB\""
		}
	}`
	var mutex = &sync.RWMutex{}
	var statusData map[string]interface{}
	json.Unmarshal([]byte(statusDataStr), &statusData)
	mutex.RLock()
	statusData["informational_settings"].(map[string]interface{})["uptime"] = uptime[botId]
	statusData["location_data"].(map[string]interface{})["position"].(map[string]interface{})["x"] = botStatus[botId].(map[string]float32)["x"]
	statusData["location_data"].(map[string]interface{})["position"].(map[string]interface{})["y"] = botStatus[botId].(map[string]float32)["y"]
	statusData["location_data"].(map[string]interface{})["position"].(map[string]interface{})["z"] = botStatus[botId].(map[string]float32)["z"]
	mutex.RUnlock()
	LAST_CLIENT_CONNECTED := "2020-08-25T06:28:44.168Z"
	if userEnv[botId] != nil {
		if userEnv[botId].(map[string]interface{})["LAST_CLIENT_CONNECTED"] != nil {
			LAST_CLIENT_CONNECTED = userEnv[botId].(map[string]interface{})["LAST_CLIENT_CONNECTED"].(string)
		}
	}
	statusData["user_env"].(map[string]interface{})["LAST_CLIENT_CONNECTED"] = LAST_CLIENT_CONNECTED

	publishGeneralData(client, botId, statusData, "status")
}
func schedulePublishStatusMessage(client mqtt_.Client, botId string) {
	async.Exec(func() interface{} {
		waitPeriod := 5
		for range time.Tick(time.Second * time.Duration(waitPeriod)) {
			publishStatusMessage(client, botId)
		}
		return nil
	})
}

/*
 * Publish telemetry data at intervals
 * exit if connection is closed
 */
func schedulePublishTelemetry(client mqtt_.Client, botId string) {
	initialTelemetryDataStr := `{
		"telemetry.captured_at": "time.Now().Format(time.RFC3339)",
		"telemetry.kind": "event",
		"telemetry.measurement": "interface_configure",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_os/platform/target/network.ex",
			"function": "{:handle_info, 2}",
			"interface": "wlan0",
			"line": 140,
			"module": "Elixir.FarmbotOS.Platform.Target.Network"
		},
		"telemetry.subsystem": "network",
		"telemetry.uuid": "9ff8173c-5fe2-4b50-b2db-da38ad52ecc4",
		"telemetry.value": null
	}`
	var initialTelemetryData map[string]interface{}

	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": "new Date().toISOString()",
		"telemetry.kind": "event",
		"telemetry.measurement": "reset",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_os/platform/target/network.ex",
			"function": "{:reset_ntp, 0}",
			"line": 358,
			"module": "Elixir.FarmbotOS.Platform.Target.Network"
		},
		"telemetry.subsystem": "ntp",
		"telemetry.uuid": "95ea2925-53e7-4e3c-a4f5-227cefe99eb6",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "interface_connect",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_os/platform/target/network.ex",
			"function": "{:handle_info, 2}",
			"interface": "wlan0",
			"line": 181,
			"module": "Elixir.FarmbotOS.Platform.Target.Network"
		},
		"telemetry.subsystem": "network",
		"telemetry.uuid": "621e5ee8-0d71-4f79-b5ec-6b67253f5908",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "lan_connect",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_os/platform/target/network.ex",
			"function": "{:handle_info, 2}",
			"interface": "wlan0",
			"line": 196,
			"module": "Elixir.FarmbotOS.Platform.Target.Network"
		},
		"telemetry.subsystem": "network",
		"telemetry.uuid": "a94eb734-bfa8-4664-8e39-e47aba534c4b",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "connection_open",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/connection_worker.ex",
			"function": "{:handle_info, 2}",
			"line": 120,
			"module": "Elixir.FarmbotExt.AMQP.ConnectionWorker"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "b0acda3b-8e27-438f-aa36-c65e08a5dc93",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "channel_open",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/log_channel.ex",
			"function": "{:handle_info, 2}",
			"line": 43,
			"module": "Elixir.FarmbotExt.AMQP.LogChannel"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "7b6b7de5-d2f7-469d-afa6-72b8cd065c1c",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "lan_connect",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_os/platform/target/network.ex",
			"function": "{:handle_info, 2}",
			"interface": "wlan0",
			"line": 196,
			"module": "Elixir.FarmbotOS.Platform.Target.Network"
		},
		"telemetry.subsystem": "network",
		"telemetry.uuid": "6d150eb5-0c03-46d9-b17d-bd668711af05",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "channel_open",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/telemetry_channel.ex",
			"function": "{:handle_info, 2}",
			"line": 58,
			"module": "Elixir.FarmbotExt.AMQP.TelemetryChannel"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "550154f8-93ef-4fb4-8efe-72fc9f9f30bb",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "channel_open",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/bot_state_channel.ex",
			"function": "{:handle_info, 2}",
			"line": 57,
			"module": "Elixir.FarmbotExt.AMQP.BotStateChannel"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "dbace292-522f-4ce5-a0d9-d70869693a1a",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry_captured_at": ""
		"telemetry_cpu_usage": 6,
		"telemetry_disk_usage": 0,
		"telemetry_memory_usage": 44,
		"telemetry_scheduler_usage": 6,
		"telemetry_soc_temp": 50,
		"telemetry_target": "rpi3",
		"telemetry_throttled": "0x0",
		"telemetry_uptime": "this.uptime",
		"telemetry_wifi_level": -39,
		"telemetry_wifi_level_percent": 90
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	initialTelemetryData["telemetry_uptime"] = uptime[botId]
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "channel_open",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/ping_pong_channel.ex",
			"function": "{:handle_info, 2}",
			"line": 73,
			"module": "Elixir.FarmbotExt.AMQP.PingPongChannel"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "fa9183a8-8829-4f67-a743-45a2bb54c38c",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "channel_open",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/celery_script_channel.ex",
			"function": "{:handle_info, 2}",
			"line": 53,
			"module": "Elixir.FarmbotExt.AMQP.CeleryScriptChannel"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "97d452cf-e39b-4deb-aad2-7256757aec9a",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "queue_bind",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/celery_script_channel.ex",
			"function": "{:handle_info, 2}",
			"line": 54,
			"module": "Elixir.FarmbotExt.AMQP.CeleryScriptChannel",
			"queue_name": "${this.botId}_from_clients",
			"routing_key": "bot.${this.botId}.from_clients"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "f5b25e2a-5593-4163-bc01-f9c34cb75cf2",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	initialTelemetryData["telemetry.meta"].(map[string]interface{})["queue_name"] = botId + "_from_clients"
	initialTelemetryData["telemetry.meta"].(map[string]interface{})["routing_key"] = botId + ".from_clients"
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "queue_bind",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/ping_pong_channel.ex",
			"function": "{:handle_info, 2}",
			"line": 75,
			"module": "Elixir.FarmbotExt.AMQP.PingPongChannel",
			"queue_name": "${this.botId}_ping",
			"routing_key": "bot.${this.botId}.ping.#"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "3ca9da32-0812-43e1-a809-482b98708562",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	initialTelemetryData["telemetry.meta"].(map[string]interface{})["queue_name"] = botId + "_png"
	initialTelemetryData["telemetry.meta"].(map[string]interface{})["routing_key"] = botId + ".ping"
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "channel_open",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/connection_worker.ex",
			"function": "{:maybe_connect, 4}",
			"line": 62,
			"module": "Elixir.FarmbotExt.AMQP.ConnectionWorker"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "2f2988a2-f290-4e1d-b396-8723cd417e81",
		"telemetry.value": null
	}`
	json.Unmarshal([]byte(initialTelemetryDataStr), &initialTelemetryData)
	initialTelemetryData["telemetry.captured_at"] = time.Now().Format(time.RFC3339)
	publishGeneralData(client, botId, initialTelemetryData, "telemetry")
	initialTelemetryDataStr = `{
		"telemetry.captured_at": ""
		"telemetry.kind": "event",
		"telemetry.measurement": "queue_bind",
		"telemetry.meta": {
			"file": "/nerves/build/farmbot_ext/lib/farmbot_ext/amqp/connection_worker.ex",
			"function": "{:maybe_connect, 4}",
			"line": 63,
			"module": "Elixir.FarmbotExt.AMQP.ConnectionWorker",
			"queue_name": "${this.botId}_auto_sync",
			"routing_key": "bot.${this.botId}.sync.#"
		},
		"telemetry.subsystem": "amqp",
		"telemetry.uuid": "4d265203-176c-4ff5-8748-4523382a6a00",
		"telemetry.value": null
	}`

	async.Exec(func() interface{} {
		waitPeriod := 1
		for range time.Tick(time.Second * time.Duration(waitPeriod)) {

			uptime[botId]++
		}
		return nil
	})

	async.Exec(func() interface{} {
		for range time.Tick(time.Second * 300) {
			telemetryData := map[string]interface{}{
				"telemetry_captured_at":        "",
				"telemetry_cpu_usage":          2,
				"telemetry_disk_usage":         0,
				"telemetry_memory_usage":       53,
				"telemetry_scheduler_usage":    2,
				"telemetry_soc_temp":           41,
				"telemetry_target":             "rpi3",
				"telemetry_throttled":          "0x0",
				"telemetry_uptime":             uptime[botId],
				"telemetry_wifi_level":         -39,
				"telemetry_wifi_level_percent": 90,
			}
			publishGeneralData(client, botId, telemetryData, "telemetry")
		}
		return nil
	})
}

func sendPingPong(botId string, client mqtt_.Client, message string) {
	client.Publish(`bot/`+botId+`/pong/`+string(message), 0, false, message)
}
