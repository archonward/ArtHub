import { API_BASE_URL } from "../../config/env";
import type { ApiErrorEnvelope, ApiResponseEnvelope } from "../../types";

export const AUTH_SESSION_EXPIRED_EVENT = "arthub:auth-session-expired";

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
  notifyOnUnauthorized?: boolean;
};

const buildRequest = (path: string, options: RequestOptions = {}) => {
  const headers = new Headers(options.headers);

  if (options.body !== undefined && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  return fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
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

    if (
      response.status === 401 &&
      options.notifyOnUnauthorized !== false &&
      typeof window !== "undefined"
    ) {
      window.dispatchEvent(
        new CustomEvent(AUTH_SESSION_EXPIRED_EVENT, {
          detail: {
            path,
            message,
            code,
          },
        }),
      );
    }

    throw new ApiError(message, response.status, code);
  }

  const payload = (await response.json()) as ApiResponseEnvelope<T>;
  return payload.data;
};
