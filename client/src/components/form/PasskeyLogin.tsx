import React, { useState } from "react";
import { startAuthentication } from "@simplewebauthn/browser";
import {
  AuthenticationResponseJSON,
  PublicKeyCredentialCreationOptionsJSON,
} from "@simplewebauthn/types";
import { Button } from "../input/Button";
import { Input } from "../input/Input";
import { AuthResponse } from "../../utils/types.ts";
import { useNavigate } from "react-router-dom";

export function PasskeyLogin(): React.ReactElement {
  const [username, setUsername] = useState("");
  const [notification, setNotification] = useState("");

  const navigate = useNavigate();

  async function loginUser() {
    if (!username) {
      setNotification("Please enter your username.");
      return;
    }

    const response = await fetch(`/login/begin`, {
      method: "POST",
      body: JSON.stringify({ username }),
      headers: {
        "Content-Type": "application/json",
      },
    });
    const credentialRequestOptions: {
      publicKey: PublicKeyCredentialCreationOptionsJSON;
    } = await response.json();
    let assertion: AuthenticationResponseJSON;
    try {
      assertion = await startAuthentication(credentialRequestOptions.publicKey);
    } catch (error) {
      if (error instanceof Error) {
        switch (error.name) {
          case "TypeError":
            setNotification(
              "There is no passkey associated with this account."
            );
            break;
          case "AbortError":
            break;
          default:
            setNotification("An error occurred. Please try again.");
        }
      }
      return;
    }

    const verificationResponse = await fetch(`/login/finish`, {
      method: "POST",
      body: JSON.stringify(assertion),
      headers: {
        "Content-Type": "application/json",
      },
    });

    const verificationJSON: AuthResponse = await verificationResponse.json();
    if (verificationJSON.status === "ok") {
      navigate("/home");
    } else {
      setNotification("Login failed.");
    }
  }

  return (
    <>
      <h2 className="text-center text-xl font-bold leading-9 tracking-tight text-gray-900">
        Sign in with passkey
      </h2>

      <div className="space-y-6">
        <div className="text-sm text-center min-h-5 font-normal text-blue-400">
          {notification}
        </div>

        <Input
          placeholder="Username"
          value={username}
          autoComplete={"webauthn"}
          onChange={setUsername}
        />

        <Button onClickFunc={loginUser} buttonText="Sign in" />
      </div>
    </>
  );
}
