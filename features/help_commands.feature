Feature: Help Commands
	Scenario: a player wants to know the commands
		Given the user has a username "server_username"
		When the user sends a "help" command
		Then the response is
		"""
```       COMMAND       |                  DESCRIPTION                   
---------------------|------------------------------------------------
    help               | Lists all the commands available for users     
    list_my_characters | Lists your characters                          
    register_character | Register your character for your discord user  
    register_player    | Ability to register yourself as player         
```
		"""
		And the response is ephemeral
