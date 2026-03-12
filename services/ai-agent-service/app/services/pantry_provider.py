class PantryProvider:
    def __init__(self):
        self.mock_pantry = [
            "butternut squash",
            "olive oil",
            "garlic",
            "salt",
            "black pepper",
            "onion",
            "vegetable broth",
            "dried thyme",
        ]

    def get_inventory(self):
        return self.mock_pantry
