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
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Type      []string  `json:"types"`
	Image     string    `json:"image"`
	Abilities []Ability `json:"abilities"` // Nouveau champ pour les capacités
}

type Ability struct {
	Name string `json:"name"`
	URL  string `json:"url"`
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

		// La fonction FetchPokemonDetails est modifiée pour retourner le nom, types, abilities, et l'image du Pokémon.
		id, name, types, abilities, image, err := FetchPokemonDetails(pokemonURL)
		if err != nil {
			log.Printf("Failed to fetch Pokemon details for ID %d: %v", pokemonID, err)
			continue // Continue to the next iteration if an error occurs
		}

		// Crée un nouvel objet Pokémon avec les détails récupérés et l'ajoute à la liste
		pokemon := Pokemon{
			ID:        id,
			Name:      name,
			Type:      types,
			Abilities: abilities, // Assurez-vous d'ajouter les capacités ici
			Image:     image,
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

func FetchPokemonDetails(pokemonURL string) (id int, name string, types []string, abilities []Ability, image string, err error) {
	resp, err := http.Get(pokemonURL)
	if err != nil {
		return 0, "", nil, nil, "", err
	}
	defer resp.Body.Close()

	var detailResp struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Sprites struct {
			Other struct {
				OfficialArtwork struct {
					FrontDefault string `json:"front_default"`
				} `json:"official-artwork"`
			} `json:"other"`
		} `json:"sprites"`
		Types []struct {
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
		} `json:"types"`
		Abilities []struct {
			Ability struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"ability"`
		} `json:"abilities"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&detailResp); err != nil {
		return 0, "", nil, nil, "", err
	}

	id = detailResp.ID
	name = detailResp.Name
	image = detailResp.Sprites.Other.OfficialArtwork.FrontDefault

	for _, t := range detailResp.Types {
		types = append(types, t.Type.Name)
	}

	for _, a := range detailResp.Abilities {
		abilities = append(abilities, Ability{
			Name: a.Ability.Name,
			URL:  a.Ability.URL,
		})
	}

	return id, name, types, abilities, image, nil
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
	id, name, types, abilities, image, err := FetchPokemonDetails(pokemonURL) // Modifié pour inclure les abilities
	if err != nil {
		log.Printf("Failed to fetch details for %s: %v", searchQuery, err) // Add logging here
		if strings.Contains(err.Error(), "404") {
			InitTemp.Temp.ExecuteTemplate(w, "search", map[string]string{
				"ErrorMessage": "Aucun Pokémon trouvé",
			})
		} else {
			http.Error(w, fmt.Sprintf("Erreur lors de la récupération des détails de Pokémon: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Si un Pokémon est trouvé, continuez comme d'habitude
	pokemon := Pokemon{
		ID:        id,
		Name:      name,
		Type:      types,
		Abilities: abilities, // Assurez-vous d'inclure les abilities ici
		Image:     image,
	}

	log.Printf("Fetched details for %s: %+v", searchQuery, pokemon)
	InitTemp.Temp.ExecuteTemplate(w, "search", pokemon)
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

func FilterPokemonParType(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erreur lors du parsing du formulaire", http.StatusBadRequest)
		return
	}

	selectedTypes := r.Form["types"]
	var filteredPokemons []Pokemon

	for _, typeName := range selectedTypes {
		pokemons, err := FetchPokemonsForType(typeName)
		if err != nil {
			log.Printf("Erreur lors de la récupération des pokémons pour le type %s: %v", typeName, err)
			continue // Passe au type suivant en cas d'erreur
		}
		filteredPokemons = append(filteredPokemons, pokemons...)
	}

	// Envoyez `filteredPokemons` à votre template HTML ici pour afficher les résultats
	InitTemp.Temp.ExecuteTemplate(w, "filtrer", filteredPokemons)
}

func PokemonDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extrayez le nom du Pokémon de l'URL
	name := strings.TrimPrefix(r.URL.Path, "/pokemon/")

	// Utilisez `FetchPokemonDetails` pour obtenir les détails du Pokémon
	id, name, types, abilities, image, err := FetchPokemonDetails("https://pokeapi.co/api/v2/pokemon/" + name) // Ajouté abilities dans la récupération
	if err != nil {
		http.Error(w, "Pokémon non trouvé", http.StatusNotFound)
		return
	}

	// Créez une instance de Pokémon avec les détails obtenus
	pokemon := Pokemon{
		ID:        id,
		Name:      name,
		Type:      types,
		Abilities: abilities, // Incluez les capacités ici
		Image:     image,
	}

	// Passez le Pokémon au template de détail
	InitTemp.Temp.ExecuteTemplate(w, "pokemon", pokemon)
}
