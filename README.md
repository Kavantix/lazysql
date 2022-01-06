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
- [x] Highlight current database
- [ ] Hosts pane
- [ ] Configuring connections from UI
- [ ] Interactive results pane
    - [X] Moving around with hjkl/arrows
    - [x] Moving around with mouse
    - [ ] Selecting rows to copy
    - [ ] Editing results
    - [ ] Resize columns manually and or to content/type
- [x] Loading indicator for the query that is running
- [ ] Resizable panes
- [ ] Query editor
    - [x] Basic VIM emulation
    - [x] Undo and Redo
    - [ ] Line numbers
    - [ ] Scrolling
    - [ ] Execute query button (instead of enter in normal mode)
    - [ ] Mouse selection
    - [ ] Basic auto complete (keywords, tables, columns)
    - [ ] Advanced auto complete (suggestions based on existing query using chroma lexer)
    - [ ] Advanced VIM emulation
- [ ] Highlight current table based on query
- [ ] Help panes like lazygit
- [ ] Searching in panes (databases, tables, results, ...)

