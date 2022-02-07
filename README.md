# Build anywhere

Python script to execute build commands (or any other commands) from anywhere in
your project as if you were in the project root directory.

## Usage

Create a `.blda` file in the root of the project. This is a JSON file containing
a map of actions to commands.

An action is passed to `blda` to have it execute the command.

``` json
// .blda
{
  "build": "cmake --build build",
  "run": "./build/executable"
}
```

``` sh
blda build # will run `cmake --build build` in the project root
blda run # will run `./build/executable` in the project root
```
