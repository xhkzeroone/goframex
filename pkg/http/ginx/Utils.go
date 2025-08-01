package ginx

func StartServer(cfg *Config, regisFunc func(ctx *RouterGroup)) error {
	srv := New(cfg)
	regisFunc(srv.rootGroup)
	// Health check endpoints
	srv.HealthCheck()

	return srv.Start()
}
