package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	InitTemp "pokemon/temp"
	"strings"
	"time"
)

var AllPokemonTypes []string

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

func init() {
	// Initialisation de AllPokemonTypes au démarrage de l'application.
	FetchPokemonTypes()
}

func FetchPokemonTypes() {
	// Effectuez la requête à l'API pour récupérer les types de Pokémon.
	resp, err := http.Get("https://pokeapi.co/api/v2/type/")
	if err != nil {
		log.Fatalf("Erreur lors de la récupération des types de Pokémon : %v", err)
	}
	defer resp.Body.Close()

	// Parsez la réponse JSON.
	var data struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatalf("Erreur lors du décodage de la réponse JSON: %v", err)
	}

	// Remplissez AllPokemonTypes avec les noms des types de Pokémon.
	for _, t := range data.Results {
		AllPokemonTypes = append(AllPokemonTypes, t.Name)
	}
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

func FetchPokemonsForType(typeName string) ([]Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/type/%s", typeName)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching type %s: %v", typeName, err)
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Pokemon []struct {
			Pokemon struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"pokemon"`
		} `json:"pokemon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var pokemons []Pokemon
	for _, p := range result.Pokemon {
		// Optionnellement, récupérez plus de détails ici avec FetchPokemonDetails
		pokemons = append(pokemons, Pokemon{Name: p.Pokemon.Name})
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

func FilterPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	InitTemp.Temp.ExecuteTemplate(w, "filter", map[string]interface{}{
		"Types": AllPokemonTypes,
	})
}

func ApplyFilterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", 400)
		return
	}

	selectedTypes := r.Form["types"]
	var filteredPokemons []Pokemon

	for _, typeName := range selectedTypes {
		pokemons, err := FetchPokemonsForType(typeName)
		if err != nil {
			log.Printf("Error fetching pokemons for type %s: %v", typeName, err)
			continue
		}
		filteredPokemons = append(filteredPokemons, pokemons...)
	}

	// Affichez les résultats
	InitTemp.Temp.ExecuteTemplate(w, "filter", map[string]interface{}{
		"Types":    AllPokemonTypes,  // Assurez-vous que cette liste est toujours passée pour reconstruire le formulaire
		"Pokemons": filteredPokemons, // Les Pokémon filtrés à afficher
	})
}
