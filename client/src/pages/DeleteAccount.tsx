import React from "react";
import { Layout } from "../components/layout/Layout";
import { Button } from "../components/input/Button.tsx";
import { Link } from "../components/input/Link.tsx";

export default function DeleteAccount(): React.ReactElement {
  return (
    <Layout title="Delete your account">
      <div>Deleting your account cannot be undone.</div>
      <div className="mb-10">
        If you're sure you want to proceed, please confirm by clicking the
        button below.
      </div>
      <Button
        onClickFunc={() => alert("Not yet implemented")}
        buttonText="Delete"
      />
      <Link href="/home" linkText="Back" />
    </Layout>
  );
}
