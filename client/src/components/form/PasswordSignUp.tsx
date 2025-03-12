import React, { useState } from "react";
import { Button } from "../input/Button";
import { Input } from "../input/Input";
import { AuthResponse } from "../../utils/types.ts";
import { useNavigate } from "react-router-dom";

export function PasswordSignUp(): React.ReactElement {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [notification, setNotification] = useState("");

  const navigate = useNavigate();

  async function registerUser() {
    if (!username) {
      setNotification("Please enter your username.");
      return;
    }

    if (password.length < 8) {
      setNotification("Password must be at least 8 characters.");
      return;
    }

    const response = await fetch(`/register/password`, {
      method: "POST",
      body: JSON.stringify({ username, password }),
      headers: {
        "Content-Type": "application/json",
      },
    });
    const registrationJSON: AuthResponse = await response.json();
    if (registrationJSON.status === "ok") {
      navigate("/home");
    } else {
      setNotification(registrationJSON.errorMessage);
    }
  }

  return (
    <>
      <h2 className="text-center text-xl font-bold leading-9 tracking-tight text-gray-900">
        Sign up using a password
      </h2>
      <div className="space-y-6">
        <div className="text-sm text-center min-h-5 font-normal text-blue-400">
          {notification}
        </div>
        <Input placeholder="Username" value={username} onChange={setUsername} />
        <Input
          type="password"
          placeholder="Password"
          value={password}
          onChange={setPassword}
        />

        <Button onClickFunc={registerUser} buttonText="Sign up" />
      </div>
    </>
  );
}
