# lazysql

A [lazygit](https://github.com/jesseduffield/lazygit) inspired terminal sql client.

## Supported databases

- [x] MySql
- [x] Postgres

## Config

The config for the client should be put in a `~/.config/lazysql/hosts.yaml` file.
An example can be found in `hosts-example.yaml`

## TODO

- [x] Show errors in popup pane
- [x] Highlight current database
- [ ] Hosts pane
- [x] Configuring connections from UI
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
    - [x] Mouse selection
    - [ ] Basic auto complete (keywords, tables, columns)
    - [ ] Advanced auto complete (suggestions based on existing query using chroma lexer)
    - [ ] Advanced VIM emulation
- [ ] Highlight current table based on query
- [ ] Help panes like lazygit
- [ ] Searching in panes (~databases~, ~tables~, results, ...)

