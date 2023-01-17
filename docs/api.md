# API endpoints

- `GET /health`
  - Checks if the service is running well
  - It doesn't accept input params
  - Result:
    - `"OK"` if the server is running well; `"KO"` otherwise
- `POST /actions/run`
  - Runs an actions
  - Params:
    - number `app_id`: the app id
    - number `user_id`: the id of the user that's requesting the execution 
      of this action
    - string `action_name`: the name of the action to be executed
    - string `action_param`: a string that's coing to be sent to the
      action as input
  - Result:
    - `result`: the return of the script's main function
    - `error`: an error message if the script failed to execute for some
      reason

# Lua SDK

This service runs Lua scripts.

Technically, it will atempt to call a `main` function that receives
one string as input and returns another string as output.

The following functions are also included for Lua scripts:

- `upload_user_file`
  - Uploads a file to a particular user's storage
  - Params:
    - string `filename`: the file name
    - string `contents`: the file's contents
  - Return: 
    - `nil` if the upload was successful, or a string detailing an error
- `upload_file`
  - Uploads a fille to an app's storage
  - Params:
    - number `user_id`: the id of user whose storage will be used
    - string `filename`: the file name
    - string `contents`: the file's contents
  - Return:
    - `nil` if the upload was sucessful, or a string detailing an error
- `download_user_file`
  - Downloads a file from a user's storage
  - Params:
    - string `filename`: the file name
  - Return:
    - The file's contents, or `nil` if it failed for some reason
- `download_file`
  - Downloads a file from the app's storage
  - Params:
    - number `user_id`: id of the user whose storage will be used
    - string `filename`: the file namee
  - Return:
    - The file's contents, or `nil` if it failed for some reason
- `check_user_file`
  - Checks if a file exists in the user's storage
  - Params:
    - string `filename`: the file name
  - Return:
    - `true` or `false` accordingly
- `check_file`
  - Checks if a file exists in the app's
  - Params:
    - number `user id`: id of the user whose storage will be checked
    - string `filename`: the file name
  - Return:
    - `true` or `false` accordingly
- `delete_user_file`
  - Deletes a user's file
  - Params:
    - string `filename`: the file name
  - Return:
    - `true` or `false` depending whether the file was deleted or not
- `delete_file`
  - Deletes an app's file
  - Params:
    - number `user_id`: id of the user whose storage will be used
    - string `filename`: the file name
  - Return:
    - `true` or `false` depending whether the file was deleted or not

