import { useContext, useRef } from 'react';
import ToggleBtn from '../utils/togglebtn';
import { UserContext } from "../contexts/usercontext";
import { useNavigate } from "react-router-dom";
import Routes from '../config';
import { ChangePassword, Login, Logout } from '../api/auth';
import { setDefaultAvatar } from '../utils/errorhandle';
import { HiCamera, HiXCircle } from 'react-icons/hi';
import SettingsItem from '../components/settings/SettingsItem';
import { UserOBJ } from '../models/models';

function Settings() {
    const naviagte = useNavigate();
    const user_ctx = useContext(UserContext);

    const password_ref = useRef<HTMLInputElement>(undefined!);
    const new_password_ref = useRef<HTMLInputElement>(undefined!);
    const confirm_password_ref = useRef<HTMLInputElement>(undefined!);

    const who_can_dm_ref = useRef<HTMLInputElement>(undefined!);

    const avatar_input = useRef<HTMLInputElement>(undefined!);
    const avatar_image = useRef<HTMLImageElement>(undefined!);

    const navigate = useNavigate();


    // TODO: почему не в слое, где все API вызовы?
    // рефакторить не рефакторить
    function avatar() {
        if (avatar_input.current.files && avatar_input.current.files.length > 0) {
            let reader = new FileReader();
            reader.readAsDataURL(avatar_input.current.files[0]);
            reader.onload = () => {
                console.log(reader.result);
                fetch(Routes.currentUser, {
                    method: "PATCH",
                    headers: {
                        "Content-Type": "application/json"
                    },
                    body: JSON.stringify({ avatar: reader.result })
                    // {avatar: wlkkr4ru8934yrfnrdh83wtn4tne8....} (и так 1 Mb ...)
                }).then(response => {
                    if (response.status === 200) {
                        // TODO: нужна зеленая плашка, popup возможно сбоку (как в аналогичном проекте с чатом)
                        alert("Successfully updated avatar!")
                        response.json().then((user: UserOBJ) => {
                            user_ctx.setAvatar(user.avatar);
                        })
                    }
                })
            }
        }
    }

    function logout() {
        Logout()
        .then(response => {
            if (response.status === 200) {
                return Login()
            } else {
                throw new Error("Logout failed")
            }
        })
        .then(response => {
            if (response.status === 200) {
                return response.json();
            } else {
                throw new Error('Login failed')
            }
        })
        .then(data => {
            if (data.redirect) {
                window.location.href = data.redirect;
            } else {
                console.error('No redirect found in response');
            }
        })
        .catch(error => {
            console.error('Error', error)
        })
    }

    function changePassword() {
        if (password_ref.current.value === "") {
            alert("Please enter your password");
            return;
        }
        if (new_password_ref.current.value === confirm_password_ref.current.value) {
            ChangePassword(password_ref.current.value, new_password_ref.current.value).then(response => {
                if (response.status === 200) {
                    alert("Successfully changed password!")
                }
            })
        } else {
            alert("New password and confirm password is not match");
            return;
        }
    }

    function onIconChange() {
        if (avatar_input.current.files && avatar_input.current.files.length > 0) {
            const file = avatar_input.current.files[0];
            console.log('avatar_input.current.value = ', avatar_input.current.value)
            if (file.size > 2097152) {
                alert("image is bigger than 2MB")
                avatar_input.current.value = ''
                return
            }
            console.log('avatar_image.current.src = ', avatar_image.current.src)
            console.log('URL.createObjectURL(file) = ', URL.createObjectURL(file))
            // файл помещается в ОЗУ браузера
            // URL.createObjectURL(file) - это временная ссылка на этот файл
            // следующая строчка позволяет отобразить данный файл 
            avatar_image.current.src = URL.createObjectURL(file);
            
        }
    }

    // TODO: картинка не обновляет при нажаатии кнопки "Save".....
    // переключаясь по вкладкам приложения не происходит запроса на @me
    // если обновить стр. он произойдет, и ава обновится



    // хэши аватарок... как это так, +1 ?...
    // 65c4e9eda7e8a0a36a7ac4d3
    // 65c4ea25a7e8a0a36a7ac4d4
    // 65c4ea40a7e8a0a36a7ac4d5

    return (
        <div className='h-full w-full overflow-hidden flex flex-col'>
            <div className='h-16 flex items-center justify-around bg-zinc-900'>
                <h2>Settings</h2>
                <button onClick={() => naviagte(-1)}><HiXCircle className='hover:text-gray-600' size={42} /></button>
            </div>
            <div className='flex flex-col h-full w-full items-center overflow-y-scroll'>
                <SettingsItem title='Profile'>
                    <div className="relative flex items-center justify-center h-32 w-32">
                        <img onClick={() => avatar_input.current.click()} src={user_ctx.avatar} onError={setDefaultAvatar} className="h-24 w-24 rounded-xl bg-zinc-900 cursor-pointer p-0 m-2 border-slate-300 border-2 border-dashed" ref={avatar_image} alt="icon" />
                        <HiCamera size={64} onClick={() => avatar_input.current.click()} className="absolute self-center justify-self-center text-white opacity-70 cursor-pointer" />
                        <input type="file" ref={avatar_input} name="filename" hidden onChange={onIconChange} accept="image/*"></input>
                    </div>
                    <input className="h-8 w-4/5 cursor-not-allowed rounded my-1 px-2 bg-zinc-700 text-gray-400" type="text" disabled value={user_ctx.username} onClick={() => alert("Username changing is not supported!")} />
                    <button className='w-24 h-10 bg-green-700 rounded hover:bg-green-800' onClick={avatar}>Save</button>
                    <div className="flex items-center space-x-2">
                        <input 
                            className="h-8 w-4/5 rounded my-1 px-2 bg-zinc-700 text-gray-400 cursor-not-allowed" 
                            type="text" 
                            disabled 
                            value={user_ctx.id} 
                        />
                        <button 
                            className='w-24 h-8 bg-yellow-700 rounded hover:bg-yellow-900' 
                            onClick={() => { navigator.clipboard.writeText(user_ctx.id) }}
                        >
                            Copy
                        </button>
                    </div>
                </SettingsItem>
                <SettingsItem title='Chanage Password'>
                    <input className="h-8 w-4/5 rounded my-1 px-2 bg-zinc-800" type="password" placeholder='Current Password' ref={password_ref} />
                    <input className="h-8 w-4/5 rounded my-1 px-2 bg-zinc-800" type="password" placeholder='New Password' ref={new_password_ref} />
                    <input className="h-8 w-4/5 rounded my-1 px-2 bg-zinc-800" type="password" placeholder='Retype New Password' ref={confirm_password_ref} />
                    <button className='w-24 h-10 bg-green-700 rounded hover:bg-green-800 my-1' onClick={changePassword}>Save</button>
                </SettingsItem>
                    {/* "Only Friends Can Dm" внутри ToggleBtn попадет через children */}
                    {/* аналог Outlet, только Outlet работает с роутерами */}
                {/* <SettingsItem title='DMs'>
                    <ToggleBtn input_ref={who_can_dm_ref}> Only Friends Can Dm </ToggleBtn>
                    <ToggleBtn input_ref={who_can_dm_ref}> Only Friends Add To Channel </ToggleBtn>
                    <button className='w-24 h-10 bg-green-700 rounded hover:bg-green-800' onClick={who_can_dm}>Save</button>
                </SettingsItem> */}
                <SettingsItem title='Logout'>
                    <button className='w-24 h-10 bg-red-800 rounded hover:bg-red-900' onClick={logout}>Logout</button>
                </SettingsItem>
            </div>
        </div>
    )
}

export default Settings;
