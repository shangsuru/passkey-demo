import React, { useState } from "react";
import { startRegistration } from "@simplewebauthn/browser";
import {
  PublicKeyCredentialCreationOptionsJSON,
  RegistrationResponseJSON,
} from "@simplewebauthn/types";
import { Button } from "../input/Button";
import { Input } from "../input/Input";
import { AuthResponse } from "../../utils/types.ts";
import { useNavigate } from "react-router-dom";

export function PasskeySignUp(): React.ReactElement {
  const [username, setUsername] = useState("");
  const [notification, setNotification] = useState("");

  const navigate = useNavigate();

  async function registerUser() {
    if (!username) {
      setNotification("Please enter your username.");
      return;
    }

    const response = await fetch(`/register/begin`, {
      method: "POST",
      body: JSON.stringify({ username }),
      headers: {
        "Content-Type": "application/json",
      },
    });
    let registrationResponse: RegistrationResponseJSON;
    try {
      const credentialCreationOptions: {
        publicKey: PublicKeyCredentialCreationOptionsJSON;
      } & AuthResponse = await response.json();
      if (credentialCreationOptions.status === "error") {
        setNotification(credentialCreationOptions.errorMessage);
        return;
      }
      registrationResponse = await startRegistration(
        credentialCreationOptions.publicKey
      );
    } catch {
      setNotification("An error occurred. Please try again.");
      return;
    }

    const verificationResponse = await fetch(`/register/finish`, {
      method: "POST",
      body: JSON.stringify(registrationResponse),
      headers: {
        "Content-Type": "application/json",
      },
    });
    const verificationJSON: AuthResponse = await verificationResponse.json();
    if (verificationJSON.status === "ok") {
      navigate("/home");
    } else {
      setNotification("Registration failed.");
    }
  }

  return (
    <>
      <h2 className="text-center text-xl font-bold leading-9 tracking-tight text-gray-900">
        Create a new account with passkey
      </h2>

      <div className="space-y-6">
        <div className="text-sm text-center min-h-5 font-normal text-blue-400">
          {notification}
        </div>

        <Input placeholder="Username" value={username} onChange={setUsername} />

        <Button onClickFunc={registerUser} buttonText="Sign up" />
      </div>
    </>
  );
}
