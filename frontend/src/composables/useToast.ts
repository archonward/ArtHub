import { readonly, ref } from "vue";

type ToastTone = "success" | "error";

type Toast = {
  id: number;
  message: string;
  tone: ToastTone;
};

const toasts = ref<Toast[]>([]);
let nextToastId = 1;

export const useToast = () => {
  const dismissToast = (id: number) => {
    toasts.value = toasts.value.filter((toast) => toast.id !== id);
  };

  const showToast = (message: string, tone: ToastTone) => {
    const id = nextToastId++;
    toasts.value = [...toasts.value, { id, message, tone }];

    window.setTimeout(() => {
      dismissToast(id);
    }, 3000);
  };

  return {
    toasts: readonly(toasts),
    showToast,
    dismissToast,
  };
};
