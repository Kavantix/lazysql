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

## TODO

- [x] Show errors in popup pane
- [ ] Highlight current database
- [ ] Hosts pane
- [ ] Configuring connections from UI
- [ ] Interactive values pane
    - [ ] Moving around with hjkl/arrows
    - [ ] Moving around with mouse
    - [ ] Editing fields
- [ ] Resizable panes
- [ ] Improved query editor
- [ ] Highlight current table based on query
