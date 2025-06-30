# License Generator API / API de Geração de Licenças

This API allows license code generation, notifying by email about license creation, approaching expiration, and expiration. It was developed using Golang and PostgreSQL.

Esta API permite gerar códigos de licença, notificando via e-mail sobre a criação, proximidade de expiração e expiração das licenças. Foi desenvolvida utilizando Golang e PostgreSQL.

---

## Technologies Used / Tecnologias Utilizadas

- **Golang**: Main language for API development.  
  Linguagem principal para desenvolvimento da API.
- **PostgreSQL**: Relational database for storing licenses.  
  Banco de dados relacional para armazenamento das licenças.
- **Gin**: Web framework for routing and middleware.  
  Framework web para roteamento e middleware.
- **Gorm**: ORM for database interactions.  
  ORM para interações com o banco de dados.
- **Gocron**: Scheduler for checking license expiration.  
  Agendador de tarefas para verificação de expiração de licenças.
- **Gomail**: Sends notification emails.  
  Envio de e-mails para notificações.
- **Swagger**: Interactive API documentation.  
  Documentação interativa da API.

---

## Installation / Instalação

1. Clone the repository / Clone o repositório:

   ```bash
   git clone https://github.com/DhioneCastilhoBarbosa/license-generator.git
   cd license-generator
   ```

2. Configure environment variables / Configure as variáveis de ambiente:

   ```bash
   cp .env.exemple .env
   # Edit the .env file with your settings / Edite o arquivo .env com suas configurações
   ```

3. Install dependencies / Instale as dependências:

   ```bash
   go mod tidy
   ```

4. Run the application / Execute a aplicação:

   ```bash
   go run main.go
   ```

---

## API Documentation / Documentação da API

Full API documentation is available via Swagger:  
A documentação completa da API está disponível via Swagger:

👉 [https://api-licenca.intelbras-cve-pro.com.br/swagger/index.html#/](https://api-licenca.intelbras-cve-pro.com.br/swagger/index.html#/)

---

## Project Structure / Estrutura do Projeto

- `controllers/`: Handles requests and responses.  
  Manipulação das requisições e respostas da API.
- `database/`: Database configuration and connection.  
  Configuração e conexão com o banco de dados.
- `docs/`: API documentation.  
  Documentação da API.
- `jobs/`: Scheduled tasks to check license expiration.  
  Tarefas agendadas para verificação de expiração de licenças.
- `middleware/`: Application middleware.  
  Middlewares utilizados na aplicação.
- `models/`: Data model definitions.  
  Definição dos modelos de dados.
- `utils/`: Utility functions, such as email sending.  
  Funções utilitárias, como envio de e-mails.

---

## Features / Funcionalidades

- Unique license code generation.  
  Geração de códigos de licença únicos.
- Email notifications for:  
  Envio de e-mails notificando:
  - License creation / Criação de nova licença.
  - Approaching expiration / Proximidade da data de expiração.
  - License expiration / Expiração da licença.
- Automatic scheduled check for expired licenses.  
  Agendamento automático de verificação de licenças expiradas.

---

## Contributing / Contribuição

Contributions are welcome! Feel free to open issues or pull requests.  
Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou pull requests.

