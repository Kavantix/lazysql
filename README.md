# lazysql

A [lazygit](https://github.com/jesseduffield/lazygit) inspired terminal sql client.

## Supported databases

- [x] MySql
- [ ] Postgres

## Config

The config for the client should be put in the `.env` file.
An example can be found in `.env.example`

### Options

| Name | Description | Default |
| :--- | :---------- | :------ |
| HOSTNAME | The host or ip where the database can be found | localhost |
| DBUSER | The username to use to login to the database | N/A |
| PASSWORD | The password to use to login to the database |  |
| PORT | The port to use to connect to the database | 3306 |

