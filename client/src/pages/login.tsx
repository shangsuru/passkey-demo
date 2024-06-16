import React, { useState } from "react";
import { startAuthentication } from "@simplewebauthn/browser";
import { AuthenticationResponseJSON } from "@simplewebauthn/types";

function Login(): React.ReactElement {
  const [username, setUsername] = useState("");
  const [notification, setNotification] = useState("");

  async function loginUser() {
    if (username === "") {
      setNotification("Please enter your username");
      return;
    }

    const response = await fetch(`/login/begin/${username}`);
    const credentialRequestOptions = await response.json();
    let assertion: AuthenticationResponseJSON;
    try {
      assertion = await startAuthentication(credentialRequestOptions.publicKey);
    } catch (error: any) {
      switch (error.name) {
        case "TypeError":
          setNotification("An account with that username does not exist.");
          break;
        default:
          setNotification("An error occurred. Please try again.");
      }
      return;
    }

    const verificationResponse = await fetch(`/login/finish/${username}`, {
      method: "POST",
      body: JSON.stringify(assertion),
      headers: {
        "Content-Type": "application/json",
      },
    });

    const verificationJSON = await verificationResponse.json();
    if (verificationJSON && verificationJSON.status === "ok") {
      setNotification("Successfully logged in.");
    } else {
      setNotification("Login failed.");
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
        autoComplete="username webauthn"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
      />
      <button onClick={loginUser}>Login</button>
      <a className="link" href="/sign-up">
        Don't have an account?
      </a>
    </>
  );
}

export default Login;
