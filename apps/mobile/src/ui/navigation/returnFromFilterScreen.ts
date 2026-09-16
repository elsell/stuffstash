type FilterReturnNavigation = {
  readonly canGoBack: () => boolean;
  readonly back: () => void;
  readonly replace: (path: '/') => void;
};

export function returnFromFilterScreen(navigation: FilterReturnNavigation): void {
  if (navigation.canGoBack()) navigation.back();
  else navigation.replace('/');
}
