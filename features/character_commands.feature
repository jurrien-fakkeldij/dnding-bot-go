
Feature: Character Commands
	Scenario: a player registers a character with a name
		Given the user has an ID "12345"
		And the user with ID "12345" is registered with the name "test_user"
		When the user sends a "register_character" command with "test_character" name as a parameter
		Then the response "test_character has been added for you" is given
		And the response is ephemeral
		And there is a character record in the database for "test_user" with the name "test_character"
