import axios from 'axios';

export interface Person {
  firstname: string;
  surname: string;
  dob: string;
  country: string;
}

// External user service configuration
const PROXY_URL = process.env.PROXY_URL;
const EXTERNAL_SERVICE_TARGET = 'http://0.0.0.0:3006';
const BASE_PROXY_URL = PROXY_URL || 'http://localhost:8099/proxy';

// Configure axios to disable compression when using proxy
const axiosConfig = PROXY_URL
  ? {
      decompress: false,
      headers: {
        'Accept-Encoding': 'identity',
      },
    }
  : {};

export class PeopleService {
  /**
   * Find person by surname and date of birth
   * Calls external user service through proxy
   * @param surname - Person's surname
   * @param dob - Date of birth in format YYYY-MM-DD
   * @returns Person object if found, null otherwise
   */
  async findPerson(surname: string, dob: string): Promise<Person | null> {
    try {
      // Build query parameters for the external service
      const queryParams = new URLSearchParams({ surname, dob }).toString();

      // Build the complete target URL (external service endpoint + query params)
      const targetEndpoint = `${EXTERNAL_SERVICE_TARGET}/person?${queryParams}`;

      // URL-encode the target for the proxy parameter
      // Use encodeURIComponent to properly encode the entire target URL
      const encodedTarget = encodeURIComponent(targetEndpoint);

      // Build the final proxy URL with the properly encoded target parameter
      const proxyUrl = `${BASE_PROXY_URL}?target=${encodedTarget}`;

      const response = await axios.get<Person>(
        proxyUrl,
        axiosConfig
      );
      return response.data;
    } catch (error: any) {
      if (error.response?.status === 404) {
        return null;
      }
      throw new Error(`Failed to fetch person: ${error.message}`);
    }
  }

  /**
   * Find people by surname OR date of birth (partial search)
   * Calls external user service through proxy
   * @param surname - Optional person's surname
   * @param dob - Optional date of birth in format YYYY-MM-DD
   * @returns Array of matching people
   */
  async findPeople(surname?: string, dob?: string): Promise<Person[]> {
    try {
      // Build query params
      const params: Record<string, string> = {};
      if (surname) params.surname = surname;
      if (dob) params.dob = dob;

      const queryParams = new URLSearchParams(params).toString();

      // Build the complete target URL (external service endpoint + query params)
      const targetEndpoint = `${EXTERNAL_SERVICE_TARGET}/person?${queryParams}`;

      // URL-encode the target for the proxy parameter
      // Use encodeURIComponent to properly encode the entire target URL
      const encodedTarget = encodeURIComponent(targetEndpoint);

      // Build the final proxy URL with the properly encoded target parameter
      const proxyUrl = `${BASE_PROXY_URL}?target=${encodedTarget}`;

      const response = await axios.get<Person[]>(
        proxyUrl,
        axiosConfig
      );
      return response.data;
    } catch (error: any) {
      throw new Error(`Failed to fetch people: ${error.message}`);
    }
  }
}
