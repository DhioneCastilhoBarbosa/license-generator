package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

var (
	SMTPServer string
	SMTPPort   string
	EmailUser  string
	EmailPass  string
)

func SetupEmailConfig() {
	SMTPServer = "smtp.gmail.com"
	SMTPPort = "587"
	EmailUser = os.Getenv("EMAILUSER")
	EmailPass = os.Getenv("EMAILPASS")
}

// EnviarEmail envia um único e-mail com múltiplos códigos de licença
func EnviarEmail(destinatario string, nome string, codigosLicenca []string, codigoCompra string) error {
	auth := smtp.PlainAuth("", EmailUser, EmailPass, SMTPServer)

	listaCodigos := strings.Join(codigosLicenca, "<br>")

	mensagem := fmt.Sprintf("To: %s\r\n"+
		"Subject: Suas Licenças Intelbras CVE-Pro foram Geradas\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		`<div style="font-family: Arial, sans-serif; background-color: #f4f4f4;">
		<div style="max-width: 600px; margin: 20px auto; background-color: #ffffff; border-radius: 8px;">
			<div style="background-color: #00a335; color: #ffffff; padding: 20px; text-align: center;">
				<h1>Licença para Plataforma CVE-Pro</h1>
			</div>
			<div style="padding: 20px; color: #333333;">
				<p>Olá, <strong>%s</strong>,</p>
				<p>Sua compra <strong>%s</strong> gerou as seguintes licenças:</p>
				<p><strong>%s</strong></p>
				<p>Adicione na plataforma CVE-Pro. Cada licença é válida para uma estação.</p>
			</div>
			<div style="background-color: #f4f4f4; color: #555555; text-align: center; padding: 10px;">
				<p>Este é um e-mail automático, por favor, não responda.</p>
				<p>Intelbras &copy; 2025</p>
			</div>
		</div>
	</div>`, destinatario, nome, codigoCompra, listaCodigos)

	err := smtp.SendMail(SMTPServer+":"+SMTPPort, auth, EmailUser, []string{destinatario}, []byte(mensagem))
	if err != nil {
		fmt.Println("Erro ao enviar e-mail:", err)
		return err
	}

	fmt.Println("E-mail enviado com sucesso para", destinatario)
	return nil
}

func EnviarAvisoRenovacao(destinatario, nome, codigo string) error {
	auth := smtp.PlainAuth("", EmailUser, EmailPass, SMTPServer)

	mensagem := fmt.Sprintf("To: %s\r\n"+
		"Subject: Sua Licença Intelbras CVE-Pro irá expirar em breve\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		`<div style="font-family: Arial, sans-serif; background-color: #f4f4f4;">
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
	</div>`, destinatario, nome, codigo)

	return smtp.SendMail(SMTPServer+":"+SMTPPort, auth, EmailUser, []string{destinatario}, []byte(mensagem))
}

func EnviarAvisoExpiracao(destinatario, nome, codigo string) error {
	auth := smtp.PlainAuth("", EmailUser, EmailPass, SMTPServer)

	mensagem := fmt.Sprintf("To: %s\r\n"+
		"Subject: Sua Licença Intelbras CVE-Pro expirou\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		`<div style="font-family: Arial, sans-serif; background-color: #f4f4f4;">
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
	</div>`, destinatario, nome, codigo)

	return smtp.SendMail(SMTPServer+":"+SMTPPort, auth, EmailUser, []string{destinatario}, []byte(mensagem))
}
