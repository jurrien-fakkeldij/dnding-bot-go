Feature: Player Commands
	Scenario: non existent player registers with a name given as a parameter
		Given the user has a username "server_username"
		When the user sends a "register_player" command with "test_user" name as a parameter
		Then the response "You have registered yourself with the name test_user" is given
		And the response is ephimeral
		And there is a player record in the database with "test_user"
	
	Scenario: non existent player registers with a name given as a parameter
		Given the user has a username "server_username"
		When the user sends a "register_player" command without a name as a parameter
		Then the response "You have registered yourself with the name server_username" is given
		And the response is ephimeral
		And there is a player record in the database with "server_username"

	Scenario: an existent player registers with the same name as a parameter
		Given the user has a username "server_username"
		And the user has an ID "12345"
		And the user with ID "12345" is registered with the name "test_user"
		When the user sends a "register_player" command with "test_user" name as a parameter
		Then the response "You already registered test_user. If this is not correct please contact the DM or admin" is given
		And the response is ephimeral
	
	Scenario: an existent player registers without a name as a parameter
		Given the user has a username "server_username"
		And the user has an ID "12345"
		And the user with ID "12345" is registered with the name "test_user"
		When the user sends a "register_player" command without a name as a parameter
		Then the response "You already registered test_user. If this is not correct please contact the DM or admin" is given
		And the response is ephimeral
