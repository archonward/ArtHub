import { CURRENT_USER_STORAGE_KEY } from "../constants/storage";
import type { User } from "../types";

export const getStoredCurrentUser = (): User | null => {
  const raw = localStorage.getItem(CURRENT_USER_STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as User;
  } catch {
    localStorage.removeItem(CURRENT_USER_STORAGE_KEY);
    return null;
  }
};

export const useCurrentUser = (): User | null => getStoredCurrentUser();
