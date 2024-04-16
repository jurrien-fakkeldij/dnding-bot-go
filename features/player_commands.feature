Feature: Player Commands
	Scenario: non existant player registers with name
		When any user sends a "register_player" command with "test_user" name
		Then a response should be given
		And the response should be "You have registered yourself with the name test_user"
		And the response should be ephimeral
