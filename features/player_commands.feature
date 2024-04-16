Feature: Player Commands
	Scenario: non existent player registers with a name given as a parameter
		Given the user has a username "server_username" on the server
		When the user sends a "register_player" command with "test_user" name as a parameter
		Then the response "You have registered yourself with the name test_user" is given
			And the response is ephimeral
	
	Scenario: non existent player registers with a name given as a parameter
		Given the user has a username "server_username" on the server
		When the user sends a "register_player" command without a name as a parameter
		Then the response "You have registered yourself with the name server_username" is given
			And the response is ephimeral
