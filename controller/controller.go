package controller

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	InitTemp "pokemon/temp"
	"time"
)

type Pokemon struct {
	Name  string   `json:"name"`
	URL   string   `json:"url"`   // Ajouté pour stocker l'URL de détail
	Type  []string `json:"types"` // Ceci sera rempli après une requête supplémentaire
	Image string   `json:"image"` // Ceci sera rempli après une requête supplémentaire
}

type ApiResponse struct {
	Count    int       `json:"count"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []Pokemon `json:"results"`
}

type Season struct {
	ID               string `json:"ID"`
	IMG              string `json:"IMG"`
	NEWPOKEMON       string `json:"NEWPOKEMON"`
	MAINCHARACTER    string `json:"MAINCHARTER"`
	LEAGUES          string `json:"LEAGUES"`
	LEGENDARYPOKEMON string `json:"LEGENDARYPOKEMON"`
}

// Supposons que cette fonction envoie une requête à l'API et récupère 20 Pokémon aléatoires
func GetRandomPokemons() ([]Pokemon, error) {
	rand.Seed(time.Now().UnixNano()) // Initialise le générateur de nombres aléatoires

	var pokemons []Pokemon

	for i := 0; i < 20; i++ {
		// Générer un ID aléatoire pour un Pokémon. Ajustez le max selon le nombre total de Pokémon disponibles.
		pokemonID := rand.Intn(898) + 1
		pokemonURL := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%d", pokemonID)

		// La fonction FetchPokemonDetails est modifiée pour retourner également le nom du Pokémon.
		name, types, image, err := FetchPokemonDetails(pokemonURL)
		if err != nil {
			return nil, err
		}

		// Crée un nouvel objet Pokémon avec les détails récupérés et l'ajoute à la liste
		pokemon := Pokemon{
			Name:  name, // Utilisez le nom réel du Pokémon retourné par FetchPokemonDetails
			Type:  types,
			Image: image,
		}
		pokemons = append(pokemons, pokemon)
	}

	return pokemons, nil
}

func FetchPokemonDetails(pokemonURL string) (name string, types []string, imageURL string, err error) {
	resp, err := http.Get(pokemonURL)
	if err != nil {
		return "", nil, "", err
	}
	defer resp.Body.Close()

	var detailResp struct {
		Name  string `json:"name"` // Ajout pour récupérer le nom
		Types []struct {
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
		} `json:"types"`
		Sprites struct {
			Other struct {
				OfficialArtwork struct {
					FrontDefault string `json:"front_default"`
				} `json:"official-artwork"`
			} `json:"other"`
		} `json:"sprites"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&detailResp); err != nil {
		return "", nil, "", err
	}

	for _, t := range detailResp.Types {
		types = append(types, t.Type.Name)
	}
	imageURL = detailResp.Sprites.Other.OfficialArtwork.FrontDefault
	name = detailResp.Name // Assigner le nom réel du Pokémon

	return name, types, imageURL, nil
}

func Index(w http.ResponseWriter, r *http.Request) {
	pokemons, err := GetRandomPokemons()
	if err != nil {
		// Gérez l'erreur, par exemple en renvoyant une erreur 500
		http.Error(w, "Erreur lors de la récupération des Pokémon", http.StatusInternalServerError)
		return
	}
	fmt.Println(pokemons)
	// Passez les Pokémon au template
	InitTemp.Temp.ExecuteTemplate(w, "index", pokemons)
}
