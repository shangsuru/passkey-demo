import React, { useEffect } from "react";
import { Layout } from "../components/layout/Layout";
import { MenuItem } from "../components/navigation/MenuItem.tsx";
import { Button } from "../components/input/Button.tsx";
import { useNavigate } from "react-router-dom";
import { AuthResponse } from "../utils/types.ts";
import { isAuthenticated, logout } from "../utils/shared.ts";

const MenuItems = [
  { title: "Change email address", link: "#" },
  { title: "Change password", link: "#" },
  { title: "Set up Two-Step Authentication", link: "#" },
  { title: "Manage Passkeys", link: "/passkeys" },
  { title: "Delete Account", link: "#" }
];

export default function Homepage(): React.ReactElement {
  const navigate = useNavigate();

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate("/");
    }
  }, []);

  async function signOut() {
    const response = await fetch("/logout", {
      method: "POST"
    });
    const data: AuthResponse = await response.json();
    if (data.status === "ok") {
      logout();
      navigate("/");
    }
  }

  return (
    <Layout>
      <div className="mb-10">
        {MenuItems.map((item) => (
          <MenuItem title={item.title} link={item.link} />
        ))}
      </div>
      <Button onClickFunc={signOut} buttonText="Sign out" />
    </Layout>
  );
}
