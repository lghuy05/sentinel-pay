import { createContext, useContext, useMemo, useRef } from "react";
import { Toast } from "primereact/toast";
import type { ToastMessage } from "primereact/toast";

interface ToastContextValue {
  show: (message: ToastMessage | ToastMessage[]) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export const ToastProvider = ({ children }: { children: React.ReactNode }) => {
  const toastRef = useRef<Toast | null>(null);

  const value = useMemo<ToastContextValue>(
    () => ({
      show: (message) => {
        toastRef.current?.show(message);
      }
    }),
    []
  );

  return (
    <ToastContext.Provider value={value}>
      {children}
      <Toast ref={toastRef} position="top-right" />
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within ToastProvider");
  }
  return context;
};
