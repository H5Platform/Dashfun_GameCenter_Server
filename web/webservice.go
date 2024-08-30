package web

import (
	"dashfun_gamecenter/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
)

type ApiModuleName string

const apiVersion = "v1"

type HttpMethod string

var once sync.Once
var inst *Service

const (
	GET    HttpMethod = "get"
	POST   HttpMethod = "post"
	PUT    HttpMethod = "put"
	DELETE HttpMethod = "delete"

	ApiModuleUser    = "user"
	ApiModuleGame    = "game"
	ApiModulePayment = "payment"
)

type ApiNode struct {
	// api所属module
	Module ApiModuleName
	// api名称，最终api地址将会是 /api/{apiVersion}/{Module}/{Name}
	Name string
	// 请求方式
	Method     HttpMethod
	Handler    func(c *gin.Context)
	Authorizer func(c *gin.Context)
}

func (n *ApiNode) ApiPath() (api string) {
	api = fmt.Sprintf("/api/%s/%s", apiVersion, n.Module)
	if len(n.Name) > 0 {
		api = api + "/" + n.Name
	}
	return
}

type ApiModule struct {
	Module ApiModuleName
	Nodes  []ApiNode
}

type Service struct {
	apiMap map[ApiModuleName]*ApiModule
}

func GetService() *Service {
	once.Do(func() {
		inst = &Service{
			apiMap: make(map[ApiModuleName]*ApiModule),
		}
	})
	return inst
}

func authMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func noCacheMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	}
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-rControl, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func (s *Service) getApiModule(name ApiModuleName) *ApiModule {
	module, ok := s.apiMap[name]
	if !ok {
		module = &ApiModule{
			Module: name,
			Nodes:  make([]ApiNode, 0),
		}
		s.apiMap[name] = module
	}
	return module
}

func (s *Service) RegisterApi(moduleName ApiModuleName, method HttpMethod, apiName string, handler func(c *gin.Context)) *ApiModule {
	module := s.getApiModule(moduleName)
	node := ApiNode{
		Module:  moduleName,
		Method:  method,
		Name:    apiName,
		Handler: handler,
	}
	module.Nodes = append(module.Nodes, node)
	log.Printf("API [%s]%s  has been registered", node.Method, node.ApiPath())
	return module
}

func (s *Service) Run() error {
	r := gin.New()
	r.Use(gin.Recovery(),
		gin.LoggerWithConfig(gin.LoggerConfig{
			Formatter: nil,
			Output:    nil,
			SkipPaths: []string{"/health"},
		}),
		noCacheMiddleWare(), CorsMiddleware() /*, authMiddleWare()*/)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	s.configSwagger(r)

	for _, module := range s.apiMap {
		for _, node := range module.Nodes {
			switch node.Method {
			case GET:
				r.GET(node.ApiPath(), node.Handler)
				break
			case POST:
				r.POST(node.ApiPath(), node.Handler)
				break
			}
		}
	}

	err := r.Run(fmt.Sprintf(":%d", config.GetConfig().Web.Port))
	//err := r.RunTLS(fmt.Sprintf(":%d", config.GetConfig().Web.Port), "./conf/server.crt", "./conf/server.key")
	return err
}

func (s *Service) configSwagger(r *gin.Engine) {
	//r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
