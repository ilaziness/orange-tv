import { create } from "zustand";

interface LoginDialogState {
  isOpen: boolean;
  open: () => void;
  close: () => void;
}

export const useLoginDialogStore = create<LoginDialogState>((set) => ({
  isOpen: false,
  open: () => set({ isOpen: true }),
  close: () => set({ isOpen: false }),
}));
