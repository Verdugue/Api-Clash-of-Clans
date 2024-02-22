package controller

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	InitTemp "pokemon/temp"
	"strings"
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

func ToLower(str string) string {
	return strings.ToLower(str)
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
		http.Error(w, "Erreur lors de la récupération des Pokémon", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Nombre de Pokémons récupérés : %d\n", len(pokemons))
	// Passez les Pokémon au template
	InitTemp.Temp.ExecuteTemplate(w, "index", pokemons)
}

func SearchPokemon(w http.ResponseWriter, r *http.Request) {
	// Extrait le terme de recherche de la requête
	searchQuery := r.URL.Query().Get("query")
	searchQuery = ToLower(searchQuery)
	if searchQuery == "" {
		http.Error(w, "Vous devez fournir un terme de recherche", http.StatusBadRequest)
		return
	}

	// Utilisez searchQuery pour faire une requête à l'API et obtenir des données sur le Pokémon
	pokemonURL := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", searchQuery)
	name, types, image, err := FetchPokemonDetails(pokemonURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors de la récupération des détails de Pokémon: %v", err), http.StatusInternalServerError)
		return
	}

	// Créez votre structure de réponse basée sur les données récupérées
	pokemon := Pokemon{
		Name:  name,
		Type:  types,
		Image: image,
	}

	// Affichez les résultats à l'aide de votre template ou retournez-les en JSON
	// Par exemple, si vous voulez juste afficher le nom du Pokémon recherché:
	// Ou si vous utilisez un template :
	InitTemp.Temp.ExecuteTemplate(w, "search", pokemon)
	// InitTemp.Temp.ExecuteTemplate(w, "search", pokemon)
}
