## File watcher


### Example for NodeJs.
 ```sh
   COMMAND_NAME="node src/index.js"
    go run main.go --path="src" --regex=".*\\.js$" --command=$COMMAND_NAME
 ```


 ### Example for calling a function from .zshrc.
 ```sh
    go run main.go --path="src" --regex=".*\\.js$" --command="listTor"
 ```