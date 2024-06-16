import Routes from "../config"

export async function Login() {
    const response = fetch(Routes.signin, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
    });
    return response;
}

export async function Logout() {
    console.log("async function login")
    try {
        const responseLogout = await fetch('/api/logout', {
            method: 'POST',
        });
        if (responseLogout.status !== 200) {
            console.error('Server responded with error:', responseLogout.status);
        } else {
            const data = await responseLogout.json();
            if (data.redirect) {
                window.location.href = data.redirect;
            }
        }
    } catch (error) {
        console.error('Error during fetch Logout:', error);
    }
}

export async function ChangePassword(current_password: string, new_password: string) {
    const response = fetch(Routes.changePassword, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            "current_password": current_password,
            "new_password": new_password
        })
    });

    return response;
}

export async function ForgotPassword(email: string) {
    const response = fetch(Routes.forgotPassword, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            "email": email
        })
    });

    return response;
}

export async function ResetPassword(token: string, password: string) {
    const response = fetch(Routes.resetPassword, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            "token": token,
            "password": password
        })
    });

    return response;
}
