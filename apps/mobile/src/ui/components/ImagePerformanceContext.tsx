import {createContext, useContext} from 'react';
import {noopPerformanceObserver} from '../../application/observability/PerformanceObserver';

const ImagePerformanceContext = createContext(noopPerformanceObserver);
export const ImagePerformanceProvider = ImagePerformanceContext.Provider;
export const useImagePerformanceObserver = () => useContext(ImagePerformanceContext);
