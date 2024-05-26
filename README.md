### Desafio Go: Sistema de Temperatura por CEP com OpenTelemetry

Este projeto consiste em dois serviços em Go:

- **Serviço A**: Recebe um CEP via POST, valida o formato e encaminha para o Serviço B.
- **Serviço B**: Recebe um CEP válido, consulta a localização através do serviço ViaCEP, e consulta a temperatura atual através do serviço WeatherAPI. Retorna a temperatura em Celsius, Fahrenheit e Kelvin, juntamente com o nome da cidade.


### Instruções para Execução

1. **Construir as Imagens**: No terminal, navegue para cada diretório `service-a` e `service-b` e execute o comando:

   ```
   docker build -t service-a .
   ```

   ```
   docker build -t service-b .
   ```

2. **Executar o Docker Compose**: Na raiz do projeto (onde está o arquivo `docker-compose.yml`), execute:

   ```
   docker-compose up
   ```


3. **Testar a Aplicação**: Após iniciar os serviços, você pode acessar:

- Para o serviço A:

   ```
   http://localhost:8080/cep
   ```
   
- Para o serviço B (substitua `CEP` por um código postal válido de 8 dígitos):
     ```
     http://localhost:8081/weather?cep=CEP
     ```

