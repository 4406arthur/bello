package core

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/4406arthur/bello/cmd/handler"
	"github.com/4406arthur/bello/pkg/stream"
	"github.com/4406arthur/bello/utils/logger"
	ginlogrus "github.com/4406arthur/gin-logrus"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

//InitRouter used config api endpoint and auth middleware
func InitRouter(log logger.Logger, config *viper.Viper) {

	r := gin.Default()
	//set up prometheus exporter
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	//inject gin middle layer logger
	r.Use(ginlogrus.Logger(log.GetLogger()))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	r.POST("/api/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	ncPool, err := stream.NewPool(config.GetString("kafka_config.host"), config.GetInt("kafka_config.subject_number"))
	if err != nil {
		log.Fatal("NA", logger.Trace(), err.Error())
	}
	subManager := stream.NewManager("voice", config.GetInt("kafka_config.conn_number"), log)
	streamHandler := handler.NewStreamHandler(ncPool, subManager, log)
	r.GET("/", streamHandler.Flow)
	//adminGroup := r.Group("/admin")
	//Token bucket: 20 tickets withun 10 sec
	//adminGroup.Use(throttle.Throttle(10, 20))
	//adminGroup.Use(RequestLogger(log))
	//ruleRepo := token.NewMongoRepository(config.GetString("mongo_config.endpoint"))
	//{
	// ruleService := token.NewRuleService(log, ruleRepo)
	// ruleHandler := apis.RuleHandlerInit(log, ruleService)
	// adminGroup.POST("/newRule", ruleHandler.NewRule)
	//adminGroup.Get("/getRule", ruleHandler.GetRule)
	//}

	if config.IsSet("server_config.cert") && config.IsSet("server_config.key") {
		r.RunTLS(
			config.GetString("server_config.listen_addr"),
			config.GetString("server_config.cert"),
			config.GetString("server_config.key"),
		)
	}

	s := &http.Server{
		Addr:           config.GetString("server_config.listen_addr"),
		Handler:        r,
		ReadTimeout:    4 * time.Second,
		WriteTimeout:   4 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}

// func RequestLogger(log logger.Logger) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		buf, _ := ioutil.ReadAll(c.Request.Body)
// 		rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
// 		rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf)) //We have to create a new Buffer, because rdr1 will be read.

// 		logger.Log(log, "Info", "RQ", logger.BuildLogInfo(c), readBody(rdr1))

// 		c.Request.Body = rdr2
// 		c.Next()
// 	}
// }

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	s := buf.String()
	return s
}
