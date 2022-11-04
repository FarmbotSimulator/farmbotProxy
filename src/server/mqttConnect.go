/* Copyright 2020 Brian Onang'o
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/csymapp/mqtt/server"
	"github.com/csymapp/mqtt/server/events"
	"github.com/csymapp/mqtt/server/listeners"
	mqtt_ "github.com/eclipse/paho.mqtt.golang"
	"github.com/logrusorgru/aurora"

	"github.com/FarmbotSimulator/farmbotProxy/config"
	async "github.com/FarmbotSimulator/farmbotProxy/src/async"
	"github.com/FarmbotSimulator/farmbotProxy/src/farmbot"
)

var uptime map[string]uint64
var userEnv map[string]interface{}
var users map[string]string
var usersOriginal map[string]string
var tokens map[string]string
var brokers map[string]string
var farmbotConnections map[string]string
var allowedTopics map[string][]string
var botStatus map[string]interface{}
var FARMBOTURL string
var server *mqtt.Server

func mqttConnect() {
	botStatus = make(map[string]interface{})
	uptime = make(map[string]uint64)
	userEnv = make(map[string]interface{})
	users = make(map[string]string)
	farmbotConnections = make(map[string]string)
	usersOriginal = make(map[string]string)
	tokens = make(map[string]string)
	brokers = make(map[string]string)
	allowedTopics = make(map[string][]string)

	FARMBOTURLInterface, _ := config.GetConfig("FARMBOTURL")
	FARMBOTURL = FARMBOTURLInterface.(string)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	fmt.Println(aurora.Magenta("Mochi MQTT Server initializing..."), aurora.Cyan("TCP"))

	server = mqtt.NewServer(nil)
	tcp := listeners.NewTCP("t1", ":1883")
	err := server.AddListener(tcp, &listeners.Config{
		// Auth: new(auth.Allow),
		Auth: &Auth{
			Users:         users,
			AllowedTopics: allowedTopics,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// stats := listeners.NewHTTPStats("stats", ":8080")
	// err = server.AddListener(stats, &listeners.Config{
	// 	Auth: new(auth.Allow),
	// 	// TLSConfig: tlsConfig,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Start the server
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	async.Exec(func() interface{} {
		var connectings = make(map[string]time.Time)
		for range time.Tick(time.Second * 5) { // try to connect new clients every 2 secs
			// loop through botIds
			// fmt.Println(tokens)

			for _, botId := range users {
				tryToConnect := true
				_, exist := connectings[botId]
				if exist {
					if (time.Now().Sub(connectings[botId]).Seconds()) >= 10 { // every 5 seconds
						tryToConnect = true
					} else {
						tryToConnect = false
					}
				}

				if farmbotConnections[botId] == botId || tryToConnect == false {
					// don't connect if there is already a connection
					// fmt.Println("connection already exists")
				} else {
					go func() {
						broker_ := brokers[botId]
						token := tokens[botId]
						broker := `wss://` + broker_ + `:443/ws/mqtt`
						opts := mqtt_.NewClientOptions().AddBroker(broker)
						// calculate the login auth info, and set it into the connection options
						opts.SetUsername(botId)
						opts.SetPassword(token)
						opts.SetKeepAlive(60 * 2 * time.Second)

						var f mqtt_.MessageHandler = func(client mqtt_.Client, msg mqtt_.Message) {
							monitorDownlinkMessages(client, msg.Topic(), msg.Payload())
						}

						opts.SetDefaultPublishHandler(f)
						var connectHandler mqtt_.OnConnectHandler = func(client mqtt_.Client) {
							farmbotConnections[botId] = botId
							if token := client.Subscribe(`bot/`+botId+`/#`, 0, nil); token.Wait() && token.Error() != nil {
								fmt.Println(token.Error())
								// os.Exit(1)
							}
							schedulePublishTelemetry(client, botId)
							schedulePublishLogs(client, botId)
							schedulePublishStatusMessage(client, botId)

							/*
								waitPeriod := 2.0
								lastPingTime := time.Now()

									async.Exec(func() interface{} {
										for range time.Tick(time.Second * time.Duration(waitPeriod)) {
											now := time.Now()
											if (now.Sub(lastPingTime).Seconds()) >= waitPeriod {
												client.Publish(`bot/`+botId+`/ping`+strconv.Itoa(int(now.Unix())), 0, false, strconv.Itoa(int(now.Unix())))
												lastPingTime = now
											}
										}
										return nil
									})
							*/
						}
						var connectLostHandler mqtt_.ConnectionLostHandler = func(client mqtt_.Client, err error) {
							delete(farmbotConnections, botId)
							delete(connectings, botId)
						}
						opts.OnConnect = connectHandler
						opts.OnConnectionLost = connectLostHandler
						client := mqtt_.NewClient(opts)
						connectings[botId] = time.Now()
						if token := client.Connect(); token.Wait() && token.Error() != nil {
							fmt.Println("FAILED...")
							// panic(token.Error())
						} else {
							fmt.Println("connected...")
							farmbotConnections[botId] = botId
						}
					}()
				}
			}
			// botId := string(cl.Username)
			// if farmbotConnections[botId] == botId || connectings[botId] == botId {
			// 	// don't connect if there is already a connection
			// 	fmt.Println("connection already exists")
			// }
		}
		return nil
	})
	// Add OnConnect Event Hook
	server.Events.OnConnect = func(cl events.Client, pk events.Packet) {
		// fmt.Printf("<< OnConnect client connected %s: %+v\n", cl.ID, pk)
		// send direct message to client
		server.Publish(string(cl.Username), []byte(users[string(cl.Username)]), false)

		//create a go routine for connecting to farmbot

	}

	// Add OnDisconnect Event Hook
	server.Events.OnDisconnect = func(cl events.Client, err error) {
		// fmt.Printf("<< OnDisconnect client disconnected %s: %v\n", cl.ID, err)
	}

	// Add OnSubscribe Event Hook
	server.Events.OnSubscribe = func(filter string, cl events.Client, qos byte) {
		server.Publish(usersOriginal[string(cl.Username)], []byte(cl.Username), false)
		// fmt.Printf("<< OnSubscribe client subscribed %s: %s %v\n", cl.ID, filter, qos)
	}

	// Add OnUnsubscribe Event Hook
	server.Events.OnUnsubscribe = func(filter string, cl events.Client) {
		// fmt.Printf("<< OnUnsubscribe client unsubscribed %s: %s\n", cl.ID, filter)
	}

	// Add OnMessage Event Hook
	server.Events.OnMessage = func(cl events.Client, pk events.Packet) (pkx events.Packet, err error) {
		// check. Work this well
		topic := pk.TopicName
		topicPart := ""
		r, _ := regexp.Compile(`^[/]` + string(cl.Username) + `/[^/]+[/]?`) // has either a slash at the end or nothing more
		for _, match := range r.FindStringSubmatch(topic) {
			match = strings.Replace(match, `/`+string(cl.Username)+"/", "", -1)
			match = strings.Replace(match, `/`, "", -1)
			topicPart = match
		}
		switch topicPart {
		case "GET":
			if resp, err := farmbot.Get(string(pk.Payload), tokens[string(cl.Username)]); err != nil {

			} else {
				// fmt.Println(resp.(string))
				server.Publish(strings.Replace(topic, "GET", fmt.Sprintf("SET/%s", string(pk.Payload)), -1), []byte(resp.(string)), false)
			}
			break
		}

		return pk, nil
	}

	// Demonstration of directly publishing messages to a topic via the
	// `server.Publish` method. Subscribe to `direct/publish` using your
	// MQTT client to see the messages.
	go func() {
		for range time.Tick(time.Second * 10) {
			server.Publish("direct/publish", []byte("scheduled message"), false)
			// fmt.Println("> issued direct message to direct/publish")
		}
	}()

	fmt.Println(aurora.BgMagenta("  Started!  "))

	<-done
	fmt.Println(aurora.BgRed("  Caught Signal  "))

	server.Close()
	fmt.Println(aurora.BgGreen("  Finished  "))
}

type Auth struct {
	Users         map[string]string   // A map of usernames (key) with passwords (value).
	AllowedTopics map[string][]string // A map of usernames and topics
}

func connectToFarmBot(email string, password string) (interface{}, error) {
	values := map[string]map[string]string{"user": {"email": email, "password": password}}
	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post(FARMBOTURL+"/tokens", "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}
	var res map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&res)

	if resp.StatusCode != 200 { // throw error
		return false, fmt.Errorf(fmt.Sprintf("%d", resp.StatusCode))
	}
	return res["token"], nil
}

func (a *Auth) Authenticate(user, password []byte) (interface{}, error) {

	var tokenInfo interface{}
	var err error
	if tokenInfo, err = connectToFarmBot(string(user), string(password)); err != nil {
		// assume webToken.
		if tokenInfo, err = connectToFarmBot(string(user), string(password)); err != nil {
			return nil, fmt.Errorf("..")
		}
	}
	token := tokenInfo.(map[string]interface{})["encoded"].(string)
	broker := tokenInfo.(map[string]interface{})["unencoded"].(map[string]interface{})["mqtt"].(string)
	botId := tokenInfo.(map[string]interface{})["unencoded"].(map[string]interface{})["bot"].(string)
	// fmt.Println(token, broker, botId)
	usersOriginal[botId] = string(user)
	users[botId] = botId
	tokens[botId] = token
	brokers[botId] = broker
	botStatus[botId] = map[string]float32{
		"x": 0, "y": 0, "z": 0,
	}
	uptime[botId] = 0
	return botId, nil
}

// ACL returns true if a user has access permissions to read or write on a topic.
func (a *Auth) ACL(user []byte, topic string, write bool) bool {
	if topic == usersOriginal[string(user)] {
		return true
	}
	r, _ := regexp.Compile(`^[/]?` + string(user) + `$`) // has either a slash at the end or nothing more
	allowed := false
	for _, _ = range r.FindStringSubmatch(topic) {
		allowed = true
	}
	r, _ = regexp.Compile(`^[/]?` + string(user) + `[/]`) // has either a slash at the end or nothing more
	for _, _ = range r.FindStringSubmatch(topic) {
		allowed = true
	}
	if allowed == false {
		r, _ := regexp.Compile(`^/simulator/` + string(user) + `/`) // has either a slash at the end or nothing more
		allowed = false
		for _, _ = range r.FindStringSubmatch(topic) {
			allowed = true
		}
		r, _ = regexp.Compile(`^/client/` + string(user) + `[/]`) // has either a slash at the end or nothing more
		for _, _ = range r.FindStringSubmatch(topic) {
			allowed = true
		}
	}
	return allowed
}
