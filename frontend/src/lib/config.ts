import { Capacitor } from "@capacitor/core";

const androidEmulatorApiUrl = "http://10.0.2.2:8080";
const browserApiUrl = "http://localhost:8080";

function getDefaultApiUrl() {
  return Capacitor.getPlatform() === "android"
    ? androidEmulatorApiUrl
    : browserApiUrl;
}

export const API_URL = import.meta.env.API_URL?.trim() || getDefaultApiUrl();
