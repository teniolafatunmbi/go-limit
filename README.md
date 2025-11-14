# go-limit (Fixed window counter algorithm)

## Implementation
- The window to window counter map is stored in-memory in the format below:
```sh
    { 
    'HH.MM' : <counter>
    }
```
The `HH.MM` is the hours and minute respectively.

- The counter threshold is set to 60
