import { API_BASE_URL } from "../../config/env";
import { CURRENT_USER_STORAGE_KEY } from "../../constants/storage";
import type { ApiErrorEnvelope, ApiResponseEnvelope } from "../../types";

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(message: string, status: number, code = "unknown_error") {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
};

const buildRequest = (path: string, options: RequestOptions = {}) => {
  const headers = new Headers(options.headers);
  const method = options.method || "GET";

  if (options.body !== undefined && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  if (method !== "GET" && method !== "HEAD" && !headers.has("X-User-ID")) {
    const raw = localStorage.getItem(CURRENT_USER_STORAGE_KEY);
    if (raw) {
      try {
        const currentUser = JSON.parse(raw) as { id?: number };
        if (currentUser.id) {
          headers.set("X-User-ID", String(currentUser.id));
        }
      } catch {
        localStorage.removeItem(CURRENT_USER_STORAGE_KEY);
      }
    }
  }

  return fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
  });
};

export const request = async <T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> => {
  const response = await buildRequest(path, options);

  if (!response.ok) {
    let message = `Request failed with status ${response.status}`;
    let code = "unknown_error";

    try {
      const payload = (await response.json()) as ApiErrorEnvelope;
      if (payload?.error?.message) {
        message = payload.error.message;
        code = payload.error.code || code;
      }
    } catch {
      const text = await response.text();
      if (text) {
        message = text;
      }
    }

    throw new ApiError(message, response.status, code);
  }

  const payload = (await response.json()) as ApiResponseEnvelope<T>;
  return payload.data;
};
