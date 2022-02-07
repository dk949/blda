# Builder. Run build scripts anywhere in the project.

Python script to execute build commands (or any other commands) from anywhere in
your project as if you were in the project root directory.

## Usage

Create a `.bldr` file in the root of the project. This is a JSON file containing
a map of actions to commands.

An action is passed to `bldr` to have it execute the command.

### Example

In .bldr

``` json
{
  "build": "cmake --build build",
  "run": "./build/executable"
}
```

``` sh
bldr build # will run `cmake --build build` in the project root
bldr run # will run `./build/executable` in the project root
```

## Build

Requires go

``` sh
make
make install
```
