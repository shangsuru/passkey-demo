import React from "react";
import { Layout } from "../components/layout/Layout";
import { LinkButton } from "../components/input/LinkButton.tsx";
import { Button } from "../components/input/Button.tsx";
import { HorizontalLine } from "../components/layout/HorizontalLine.tsx";
import { Link } from "../components/input/Link.tsx";

export default function ManagePasskeys(): React.ReactElement {
  const registeredPasskeys = [
    {
      name: "Chrome on Mac",
      registeredAt: "July 19, 2024",
      lastUsedAt: "July 19, 2024",
    },
    {
      name: "Edge on Windows",
      registeredAt: "July 17, 2024",
      lastUsedAt: "July 19, 2024",
    },
  ];

  return (
    <Layout title="Manage Passkeys">
      <Button
        onClickFunc={() => alert("Not implemented yet!")}
        buttonText="Register a passkey"
      />
      <div className="font-light text-xs mt-2">
        A prompt will be displayed to confirm registration.
      </div>
      <HorizontalLine />
      {registeredPasskeys.map((passkey) => (
        <div>
          <div key={passkey.name} className="grid grid-cols-2 gap-4">
            <div>
              <div className="font-bold">{passkey.name}</div>
              <div className="font-light text-xs text-gray-400">
                <p>Registered at: {passkey.registeredAt}</p>
                <p>Last used at: {passkey.lastUsedAt}</p>
              </div>
            </div>
            <div>
              <LinkButton
                onClickFunc={() => alert("Not implemented yet!")}
                buttonText="Delete"
              />
            </div>
          </div>
          <HorizontalLine />
        </div>
      ))}
      <Link href="/home" linkText="Back" />
    </Layout>
  );
}
