package serverimpl

import (
	"FantasticLife/server/serverimpl/WebSocket"
	"FantasticLife/server/serverimpl/cache"
	"FantasticLife/server/serverimpl/mysqlCli"
	"FantasticLife/server/serverimpl/task"
	"go.uber.org/fx"
)

var Module = fx.Module("serverimpl",
	fx.Provide(NewLLMBOT),
	fx.Provide(NewLLMTransceiver),
	fx.Provide(WebSocket.NewClientManager),
	fx.Provide(cache.NewRedisCli),
	fx.Provide(mysqlCli.NewMysqlCli),
	fx.Invoke(task.Init),
	fx.Invoke(task.ServerInit),
)
