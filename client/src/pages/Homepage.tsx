import React from "react";
import { Layout } from "../components/layout/Layout";
import { MenuItem } from "../components/navigation/MenuItem.tsx";
import { Button } from "../components/input/Button.tsx";

const MenuItems = [
  { title: "Manage Passkeys", link: "/passkeys" },
  { title: "Delete Account", link: "/delete_account" },
];

export default function Homepage(): React.ReactElement {
  async function signOut() {
    await fetch("/logout", {
      method: "POST",
    });

    window.location.reload();
  }

  return (
    <Layout title="Homepage">
      <div className="mb-10">
        {MenuItems.map((item) => (
          <MenuItem title={item.title} link={item.link} />
        ))}
      </div>
      <Button onClickFunc={signOut} buttonText="Sign out" />
    </Layout>
  );
}
