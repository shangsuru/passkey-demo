import React, { useState } from "react";
import { startRegistration } from "@simplewebauthn/browser";
import { RegistrationResponseJSON } from "@simplewebauthn/types";

function Register(): React.ReactElement {
  const [username, setUsername] = useState("");
  const [notification, setNotification] = useState("");

  async function registerUser() {
    if (username === "") {
      setNotification("Please enter your username");
      return;
    }

    const response = await fetch(`/register/begin/${username}`);
    let registrationResponse: RegistrationResponseJSON;
    try {
      const credentialCreationOptions = await response.json();
      registrationResponse = await startRegistration(
        credentialCreationOptions.publicKey,
      );
    } catch (error: any) {
      switch (error.name) {
        case "InvalidStateError":
          setNotification("An account with that username already exists.");
          break;
        default:
          setNotification("An error occurred. Please try again.");
      }
      return;
    }

    const verificationResponse = await fetch(`/register/finish/${username}`, {
      method: "POST",
      body: JSON.stringify(registrationResponse),
      headers: {
        "Content-Type": "application/json",
      },
    });
    const verificationJSON = await verificationResponse.json();

    if (verificationJSON && verificationJSON.status === "ok") {
      setNotification("Successfully registered.");
    } else {
      setNotification("Registration failed.");
    }
  }

  return (
    <>
      <div className="header">
        <h1>Passkey Demo</h1>
      </div>
      <div id="notification">{notification}</div>
      <input
        type="text"
        name="username"
        id="username"
        placeholder="Username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
      />
      <button onClick={registerUser}>Sign-up</button>
      <a className="link" href="/">
        Already have an account?
      </a>
    </>
  );
}

export default Register;
