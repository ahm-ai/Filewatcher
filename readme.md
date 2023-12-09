## File watcher


### Example for NodeJs.
 ```sh

   COMMAND_NAME="
     echo 'Multi'; \
     echo 'Line'; \
     listTor 
   "

    go run main.go --path="<ABSOLUTE_PATH>" --regex=".*\\.js$" --command=$COMMAND_NAME
 ```


 ### Example for calling a function from .zshrc
 ```sh

   COMMAND_NAME="
      node src/index.js
   "

    go run main.go --path="src" --regex=".*\\.js$" --command=$COMMAND_NAME
 ```


### Example when added to .vscode folder
 ```sh
   COMMAND_NAME="echo 'hi'"
    go run main.go --path="../" --regex=".*\\.js$" --command=$COMMAND_NAME
 ```