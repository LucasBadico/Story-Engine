import { createContext, useContext, PropsWithChildren } from 'react';
import type { Services } from '@story-engine/shared-ts';

const ServicesContext = createContext<Services | null>(null);

export interface ServicesProviderProps {
  services: Services;
}

export function ServicesProvider({ services, children }: PropsWithChildren<ServicesProviderProps>) {
  return (
    <ServicesContext.Provider value={services}>
      {children}
    </ServicesContext.Provider>
  );
}

export function useServices(): Services {
  const context = useContext(ServicesContext);
  if (!context) {
    throw new Error('useServices must be used within ServicesProvider');
  }
  return context;
}

