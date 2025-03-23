import axios from "axios";

const API_URL = (function () {
    // Ottieni l'host dal sito web corrente
    const host = document.location.protocol + "//" + document.location.hostname;

    // Estrai la porta dall'API definita in __API_URL__
    const port = __API_URL__.split(":").pop();

    // Componi e restituisci l'URL dell'API
    return `${host}:${port}`;
})();

const instance = axios.create({
    baseURL: API_URL,
    timeout: 1000 * 5
});

export function getUrl (variable) {
    console.log(variable);
    return `${API_URL}/${variable}`;
}

export default instance;

