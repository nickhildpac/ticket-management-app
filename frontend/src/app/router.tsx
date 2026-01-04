import { createRootRouteWithContext, createRoute, createRouter, Outlet, redirect } from '@tanstack/react-router';
import { QueryClient } from '@tanstack/react-query';
import { Dashboard } from "@/features/dashboard";
import { Login } from "@/features/auth/login";
import { TicketList } from "@/features/tickets/list";
import { TicketDetails } from "@/features/tickets/details";
import { TicketForm } from "@/features/tickets/form";
import { AdminPanel } from "@/features/admin";
import { isAuthenticated } from "@/app/auth";

export type AppContext = { qc: QueryClient; };

const rootRoute = createRootRouteWithContext<AppContext>()({
    component: () => <Outlet />,
});

// Auth Guard Helpers
const requireAuth = () => {
    if (!isAuthenticated()) {
        throw redirect({ to: '/login' });
    }
};

const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: Dashboard,
    beforeLoad: requireAuth
});

const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/login',
    component: Login,
});

const ticketsListRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets',
    component: TicketList,
    beforeLoad: requireAuth
});

const ticketCreateRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets/new',
    component: TicketForm,
    beforeLoad: requireAuth // Should check role too
});

const ticketDetailsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets/$id',
    component: TicketDetails,
    beforeLoad: requireAuth
});

const adminRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/admin',
    component: AdminPanel,
    beforeLoad: requireAuth // Should check admin role
});

const routeTree = rootRoute.addChildren([
    indexRoute,
    loginRoute,
    ticketsListRoute,
    ticketCreateRoute,
    ticketDetailsRoute,
    adminRoute,
]);

export const router = createRouter({
    routeTree,
    context: { qc: new QueryClient() } // Placeholder qc, real one passed in main
});

declare module '@tanstack/react-router' {
    interface Register {
        router: typeof router
    }
}
