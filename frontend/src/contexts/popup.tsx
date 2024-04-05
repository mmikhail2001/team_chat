import { useState, createContext } from "react";

export interface PopUpContextOBJ {
    component: React.ReactChild | undefined
    show: boolean
    open: (child: React.ReactChild) => void
    close: () => void
}

export const PopUpContext = createContext<PopUpContextOBJ>(undefined!);

// вот так делаются модальные окна
// https://github.com/DmitriyShkredov/react-simple-modal
// не понимаю close, open как влияют


// TODO: PopUpCTX - единообразие (ChannelCTX)
export default function PopUpProvider({ children }: { children: React.ReactChild }) {
    const [component, setComponent] = useState<React.ReactChild>();
    const [show, setShow] = useState<boolean>(false);

    const open = (child: React.ReactChild) => {
        setComponent(child);
        setShow(true)
    }

    const close = () => {
        setShow(false)
        setComponent(undefined);
    }

    const value: PopUpContextOBJ = {
        component,
        show,
        open,
        close
    }

    return (
        <PopUpContext.Provider value={value}>
            {children}
        </PopUpContext.Provider>
    );
}
