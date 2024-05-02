Feature: Help Commands
	Scenario: a player wants to know the commands
		Given the user has a username "server_username"
		When the user sends a "help" command
		Then the response "You have registered yourself with the name test_user" is given
		And the response is ephimeral
		And there is a player record in the database with "test_user"
