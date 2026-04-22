package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

var (
	SMTPServer = "postal.intelbras.com.br"
	SMTPPort   = 2525
	EmailUser  = "" // remover e verificar
	EmailPass  = ""
)

// SetupEmailConfig carrega usuário e senha do .env
func SetupEmailConfig() {
	EmailUser = os.Getenv("EMAILUSER")
	EmailPass = os.Getenv("EMAILPASS")
}

// sendEmail envia um e-mail usando go-simple-mail
func sendEmail(to string, subject string, body string) error {
	server := mail.NewSMTPClient()
	server.Host = SMTPServer
	server.Port = SMTPPort
	server.Username = EmailUser
	server.Password = EmailPass
	server.Encryption = mail.EncryptionNone
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor SMTP: %w", err)
	}

	email := mail.NewMSG()
	email.SetFrom(fmt.Sprintf("%s <%s>", "Intelbras CVE", "licenca.cve@intelbras.com.br")).
		AddTo(to).
		SetSubject(subject).
		SetBody(mail.TextHTML, body)

	if email.Error != nil {
		return fmt.Errorf("erro ao criar e-mail: %w", email.Error)
	}

	if err := email.Send(smtpClient); err != nil {
		return fmt.Errorf("erro ao enviar e-mail: %w", err)
	}

	fmt.Println("E-mail enviado com sucesso para", to)
	return nil
}

// EnviarEmail envia o e-mail com as licenças
func EnviarEmail(destinatario string, nome string, codigosLicenca []string, codigoCompra string) error {
	listaCodigos := strings.Join(codigosLicenca, "<br>")

	body := fmt.Sprintf(`<div style="font-family: Arial, sans-serif; background-color: #f4f4f4;">
	<div style="max-width: 600px; margin: 20px auto; background-color: #ffffff; border-radius: 10px; overflow: hidden;">
		
		<div style="background-color: #00a335; color: #ffffff; padding: 20px; text-align: center;">
			<h1 style="margin:0;">Licença para Plataforma CVE-Pro</h1>
		</div>

		<div style="padding: 24px; color: #333333; font-size:14px; line-height:1.6;">
			
			<p>Olá, <strong>%s</strong>,</p>

			<p>Sua compra <strong>%s</strong> gerou as seguintes licenças:</p>

			<div style="background:#f9f9f9; padding:12px; border-radius:6px; margin:10px 0;">
				<strong>%s</strong>
			</div>

			<p>
				Cada licença é válida para <strong>uma estação de recarga de veículos elétricos</strong>.
			</p>

			<p>
				<strong>Atenção:</strong> os passos abaixo são necessários apenas se você ainda não possui uma conta na plataforma CVE-Pro.
			</p>

			<p style="margin-top:20px;"><strong>Passo 1:</strong> Solicitar a chave de acesso</p>

			<a href="https://license.intelbras-cve-pro.com.br/solicitar-chave"
			   style="display:block; width:100%%; max-width:280px; margin:10px auto; text-align:center; background:#00a335; color:#ffffff; padding:14px 0; text-decoration:none; border-radius:6px; font-weight:bold;">
			   1. Solicitar chave de acesso
			</a>

			<p style="margin-top:20px;"><strong>Passo 2:</strong> Criar sua conta com a chave recebida</p>

			<a href="https://intelbras-cve-pro.com.br/auth/createaccount"
			   style="display:block; width:100%%; max-width:280px; margin:10px auto; text-align:center; background:#007bff; color:#ffffff; padding:14px 0; text-decoration:none; border-radius:6px; font-weight:bold;">
			   2. Criar conta com a chave
			</a>

			<p style="margin-top:20px;">
				Se você já possui uma conta, basta acessar a plataforma e adicionar suas licenças.
			</p>

		</div>

		<div style="background-color: #f4f4f4; color: #555555; text-align: center; padding: 10px;">
			<p style="margin:5px 0;">Este é um e-mail automático, por favor, não responda.</p>
			<p style="margin:5px 0;">Intelbras &copy; 2025</p>
		</div>

	</div>
</div>`, nome, codigoCompra, listaCodigos)

	return sendEmail(destinatario, "Suas Licenças Intelbras CVE-Pro foram Geradas", body)
}

// EnviarAvisoRenovacao envia e-mail de aviso de renovação
func EnviarAvisoRenovacao(destinatario, nome, codigo string) error {
	body := fmt.Sprintf(`<div style="font-family: Arial, sans-serif; background-color: #f4f4f4;">
	<div style="max-width: 600px; margin: 20px auto; background-color: #ffffff; border-radius: 8px;">
		<div style="background-color: #00a335; color: #ffffff; padding: 20px; text-align: center;">
			<h1>Licença Plataforma CVE-Pro</h1>
		</div>
		<div style="padding: 20px; color: #333333;">
			<p>Olá, <strong>%s</strong>,</p>
			<p>Atenção! Sua licença <strong>%s</strong> irá expirar em 3 dias.</p>
			<p>Considere renová-la para continuar utilizando a plataforma CVE-Pro sem interrupções.</p>
		</div>
		<div style="background-color: #f4f4f4; color: #555555; text-align: center; padding: 10px;">
			<p>Este é um e-mail automático, por favor, não responda.</p>
			<p>Intelbras &copy; 2025</p>
		</div>
	</div>
</div>`, nome, codigo)

	return sendEmail(destinatario, "Sua Licença Intelbras CVE-Pro irá expirar em breve", body)
}

// EnviarAvisoExpiracao envia e-mail de expiração
func EnviarAvisoExpiracao(destinatario, nome, codigo string) error {
	body := fmt.Sprintf(`<div style="font-family: Arial, sans-serif; background-color: #f4f4f4;">
	<div style="max-width: 600px; margin: 20px auto; background-color: #ffffff; border-radius: 8px;">
		<div style="background-color: #00a335; color: #ffffff; padding: 20px; text-align: center;">
			<h1>Licença Plataforma CVE-Pro</h1>
		</div>
		<div style="padding: 20px; color: #333333;">
			<p>Olá, <strong>%s</strong>,</p>
			<p>Atenção! Sua licença <strong>%s</strong> expirou.</p>
			<p>Para continuar utilizando a plataforma CVE-Pro, será necessário renovar sua licença.</p>
		</div>
		<div style="background-color: #f4f4f4; color: #555555; text-align: center; padding: 10px;">
			<p>Este é um e-mail automático, por favor, não responda.</p>
			<p>Intelbras &copy; 2025</p>
		</div>
	</div>
</div>`, nome, codigo)

	return sendEmail(destinatario, "Sua Licença Intelbras CVE-Pro expirou", body)
}

func EnviarEmailChave(destinatario, nome, chave string) error {
	body := fmt.Sprintf(`<div style="font-family: Arial, sans-serif; background-color: #f4f4f4;">
	<div style="max-width: 600px; margin: 20px auto; background-color: #ffffff; border-radius: 10px; overflow:hidden;">
		
		<div style="background-color: #00a335; color: #ffffff; padding: 20px; text-align: center;">
			<h1 style="margin:0;">Chave de Acesso - CVE-Pro</h1>
		</div>

		<div style="padding: 24px; color: #333333; font-size:14px; line-height:1.6;">
			
			<p>Olá, <strong>%s</strong>,</p>

			<p>Utilize a chave abaixo para criar sua conta na plataforma:</p>

			<div style="background:#f9f9f9; padding:16px; border-radius:6px; text-align:center; margin:15px 0; font-size:18px;">
				<strong>%s</strong>
			</div>

			<p>Após copiar a chave, clique no botão abaixo para criar sua conta.</p>

			<a href="https://intelbras-cve-pro.com.br/auth/createaccount"
			   style="display:block; width:100%%; max-width:280px; margin:20px auto; text-align:center; background:#007bff; color:#ffffff; padding:14px 0; text-decoration:none; border-radius:6px; font-weight:bold;">
			   Criar conta com esta chave
			</a>

		</div>

		<div style="background-color: #f4f4f4; color: #555555; text-align: center; padding: 10px;">
			<p style="margin:5px 0;">Este é um e-mail automático, por favor, não responda.</p>
			<p style="margin:5px 0;">Intelbras &copy; 2025</p>
		</div>

	</div>
</div>`, nome, chave)

	return sendEmail(destinatario, "Sua Chave de Acesso a Plataforma Intelbras CVE", body)
}
