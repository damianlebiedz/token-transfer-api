# Token Transfer GraphQL API

This project implements the GraphQL API for transferring BTP tokens between wallets. The API enables the transfer of tokens from one wallet to another, providing error handling in case of insufficient balance or race conditions.
- [Technologies](#technologies)
- [Installation](#installation)
- [Usage](#usage)
- [Tests](#tests)
- [Contact](#contact)

## Technologies

- Golang 1.24.2
- GraphQL
- Docker Compose
- PostgreSQL
- GORM
- tested with testify

## Installation

1. Download Docker Compose following instructions: https://docs.docker.com/compose/install/

2. Clone the repository:

```
git clone https://github.com/damianlebiedz/token-transfer-api.git
cd token-transfer-api
```

3. Create a `.env` file in the project's directory with the following variables:

```
# Production database configuration

POSTGRES_USER=your_user
POSTGRES_PASSWORD=your_password

# Test database configuration

TEST_POSTGRES_USER=your_test_user
TEST_POSTGRES_PASSWORD=your_test_password
```

> [!IMPORTANT]
> `.env` file is required to run the application. Never commit your .env file to a public repository!

## Usage

1. Build and run the application using docker-compose:

```
docker-compose up --build app
```

2. Access the GraphQL playground at: http://localhost:8080/playground
This is where you can test the GraphQL queries and mutations.

Example mutation:
```
mutation {
  transfer(from: "0x0000000000000000000000000000000000000000", to: "0x0000000000000000000000000000000000000001", amount: 100) {
    balance
  }
}
```
This mutation creates a second wallet and transfers a balance of 100 tokens from the initial wallet (0x000...0000) to the new one (0x000...0001). It returns the updated balance of the sender wallet after the transfer.

## Tests

- Build and run tests using docker-compose:
```
docker-compose up --build test
```

> [!NOTE]
> You can configure your own tests in `transfer_test.go` file.

## Contact
Damian Lebiedź | https://damianlebiedz.github.io/contact.html