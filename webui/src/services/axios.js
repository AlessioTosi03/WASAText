import axios from "axios";

const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});

export function getUrl (variable) {
	console.log(variable);
	return `${__API_URL__}/${variable}`;
}

export default instance;
