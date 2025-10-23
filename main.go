package main

import (
	"cve-pro-license-api/controllers"
	"cve-pro-license-api/database"
	"cve-pro-license-api/jobs"
	"cve-pro-license-api/middleware"
	"cve-pro-license-api/models"
	"cve-pro-license-api/utils"
	"os"

	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	_ "cve-pro-license-api/docs" // Importação dos documentos gerados

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // Swaggo
)

// @title API de Licenças
// @version 1.0
// @description API para gerenciar licenças de software.
// @host api-licenca.intelbras-cve-pro.com.br
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @type apiKey
// @in header
// @name Authorization
// @description Insira o token no formato `Bearer {seu_token}`
func main() {

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("Arquivo .env não encontrado, usando variáveis do ambiente")
		}
	}
	controllers.StartOrderWorker() // Inicia o worker para processar pedidos VTEX
	utils.SetupEmailConfig()
	// Inicializa o banco de dados
	database.Conectar()
	database.DB.AutoMigrate(&models.License{}, &models.Usuario{}, &models.Chave{})

	c := cron.New()

	// Executa a verificação todo dia às 02:00 da manhã
	_, err := c.AddFunc("0 1 * * *", func() {
		//_, err = c.AddFunc("@every 1m", func() { // Para teste, executa a cada 1 minuto
		log.Println("Iniciando verificação de licenças expiradas...")
		jobs.VerificarLicencasExpiradas()
	})
	if err != nil {
		log.Fatalf("Erro ao agendar tarefa: %v", err)
	}

	c.Start()

	r := gin.Default()
	// Configuração do CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // permite qualquer origem
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))
	// Rota para documentação Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("https://api-licenca.intelbras-cve-pro.com.br/swagger/doc.json")))

	// Rotas para autenticação
	r.POST("/cadastrar-usuario", controllers.CadastrarUsuario)
	r.POST("/login", controllers.Login)
	r.POST("/webhook/vtex-vendas", controllers.VtexWebhook)
	r.POST("/criar-chave", controllers.CriarChave)
	r.GET("/recuperar-chave", controllers.RecuperarChaves)

	// Rotas protegidas com autenticação
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/criar-licenca", controllers.CriarLicenca)
		protected.PUT("/atualizar-licenca", controllers.AtualizarStatusLicenca)
		protected.GET("/licencas", controllers.ListarLicencas)
		protected.GET("/chaves", controllers.ListarChaves)
		protected.PUT("/atualizar-status-chave", controllers.AtualizarStatusChave)
		protected.GET("/buscar-chave", controllers.BuscarChave)
	}

	r.Run(":8085") // Inicia o servidor na porta 8085
}
