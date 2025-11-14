# go-limit (Fixed window counter algorithm)

## TODO
- Write tests

## Implementation
- The window to window-counter map is stored in-memory in the format below:
```sh
    { 
    'HH.MM' : <counter>
    }
```
The `HH.MM` is the hours and minute respectively.

- The counter threshold is set to 60

- All data is stored in memory. If the application restarts, the window to window-counter map is reset.
