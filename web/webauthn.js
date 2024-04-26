// Base64 to ArrayBuffer
function bufferDecode(value) {
  value = value.replace(/-/g, "+").replace(/_/g, "/");
  return Uint8Array.from(atob(value), (c) => c.charCodeAt(0));
}

// ArrayBuffer to URLBase64
function bufferEncode(value) {
  return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

function registerUser() {
  const username = document.getElementById("username").value;
  if (username === "") {
    alert("Please enter your username");
    return;
  }

  fetch(`/register/begin/${username}`)
    .then(function (response) {
      return response.json();
    })
    .then((credentialCreationOptions) => {
      credentialCreationOptions.publicKey.challenge = bufferDecode(
        credentialCreationOptions.publicKey.challenge,
      );
      credentialCreationOptions.publicKey.user.id = bufferDecode(
        credentialCreationOptions.publicKey.user.id,
      );

      if (credentialCreationOptions.publicKey.excludeCredentials) {
        credentialCreationOptions.publicKey.excludeCredentials.forEach(
          (item) => {
            item.id = bufferDecode(item.id);
          },
        );
      }

      return navigator.credentials.create({
        publicKey: credentialCreationOptions.publicKey,
      });
    })
    .then((credential) => {
      let attestationObject = credential.response.attestationObject;
      let clientDataJSON = credential.response.clientDataJSON;
      let rawId = credential.rawId;

      fetch(`/register/finish/${username}`, {
        method: "POST",
        body: JSON.stringify({
          id: credential.id,
          rawId: bufferEncode(rawId),
          type: credential.type,
          response: {
            attestationObject: bufferEncode(attestationObject),
            clientDataJSON: bufferEncode(clientDataJSON),
          },
        }),
        headers: {
          "Content-Type": "application/json",
        },
      });
    })
    .then(() => {
      document.getElementById("notification").innerHTML =
        "Successfully registered.";
    })
    .catch((err) => {
      console.warn(err);
      document.getElementById("notification").innerHTML =
        "Registration failed.";
    });
}

function loginUser() {
  const username = document.getElementById("username").value;
  if (username === "") {
    alert("Please enter your username");
    return;
  }

  fetch(`/login/begin/${username}`)
    .then(function (response) {
      return response.json();
    })
    .then((credentialRequestOptions) => {
      credentialRequestOptions.publicKey.challenge = bufferDecode(
        credentialRequestOptions.publicKey.challenge,
      );
      credentialRequestOptions.publicKey.allowCredentials.forEach(
        function (listItem) {
          listItem.id = bufferDecode(listItem.id);
        },
      );

      return navigator.credentials.get({
        publicKey: credentialRequestOptions.publicKey,
      });
    })
    .then((assertion) => {
      let authData = assertion.response.authenticatorData;
      let clientDataJSON = assertion.response.clientDataJSON;
      let rawId = assertion.rawId;
      let sig = assertion.response.signature;
      let userHandle = assertion.response.userHandle;

      fetch(`/login/finish/${username}`, {
        method: "POST",
        body: JSON.stringify({
          id: assertion.id,
          rawId: bufferEncode(rawId),
          type: assertion.type,
          response: {
            authenticatorData: bufferEncode(authData),
            clientDataJSON: bufferEncode(clientDataJSON),
            signature: bufferEncode(sig),
            userHandle: bufferEncode(userHandle),
          },
        }),
        headers: {
          "Content-Type": "application/json",
        },
      });
    })
    .then(() => {
      document.getElementById("notification").innerHTML =
        "Successfully logged in.";
    })
    .catch((error) => {
      console.log(error);
      document.getElementById("notification").innerHTML = "Login failed.";
    });
}
