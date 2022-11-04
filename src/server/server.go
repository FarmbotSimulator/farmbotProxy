package server

// _ "github.com/FarmbotSimulator/farmbotProxy/docs"

func Start(production bool) {
	startMQTT()
	// env := "dev"
	// if production {
	// 	env = "prod"
	// }

	// port, err := config.GetConfig("PORT", env)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// portStr := fmt.Sprintf("%v", port)
	// portStr = fmt.Sprintf(":%s", portStr)
	// db.Connect()
	// app := fiber.New()

	// app.Use(cors.New(cors.Config{
	// 	AllowCredentials: true,
	// }))

	// routes.Setup(app)
	// if err = app.Listen(portStr); err != nil {
	// 	log.Fatal(err)
	// }
}

func startMQTT() {
	mqttConnect()
	// Exec(func() interface{} {
	// 	mqttConnect()
	// 	return nil
	// })
}
