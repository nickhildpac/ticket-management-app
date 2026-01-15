import { createRootRouteWithContext, createRoute, createRouter, Outlet, redirect } from '@tanstack/react-router';
import { QueryClient } from '@tanstack/react-query';
import { Dashboard } from "@/features/dashboard";
import { Login } from "@/features/auth/login";
import { Signup } from "@/features/auth/signup";
import { TicketList } from "@/features/tickets/list";
import { AllTicketsList } from "@/features/tickets/all-tickets";
import { AssignedTicketsList } from "@/features/tickets/assigned-tickets";
import { TicketDetails } from "@/features/tickets/details";
import { TicketForm } from "@/features/tickets/form";
import { AdminPanel } from "@/features/admin";
import { Profile } from "@/features/profile/profile";
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

const requireAdmin = () => {
    if (!isAuthenticated()) {
        throw redirect({ to: '/login' });
    }
    const userStr = localStorage.getItem('user');
    if (!userStr) {
        throw redirect({ to: '/login' });
    }
    const user = JSON.parse(userStr);
    if (user.role !== 'admin') {
        throw redirect({ to: '/' });
    }
};

const requireAgent = () => {
    if (!isAuthenticated()) {
        throw redirect({ to: '/login' });
    }
    const userStr = localStorage.getItem('user');
    if (!userStr) {
        throw redirect({ to: '/login' });
    }
    const user = JSON.parse(userStr);
    if (user.role !== 'agent' && user.role !== 'admin') {
        throw redirect({ to: '/' });
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

const signupRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/signup',
    component: Signup,
});

const ticketsListRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets',
    component: TicketList,
    beforeLoad: requireAuth
});

const ticketsAllRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets/all',
    component: AllTicketsList,
    beforeLoad: requireAdmin
});

const ticketsAssignedRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/tickets/assigned',
    component: AssignedTicketsList,
    beforeLoad: requireAgent
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
    beforeLoad: requireAdmin
});

const profileRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/profile',
    component: Profile,
    beforeLoad: requireAuth
});

const routeTree = rootRoute.addChildren([
    indexRoute,
    loginRoute,
    signupRoute,
    ticketsListRoute,
    ticketsAllRoute,
    ticketsAssignedRoute,
    ticketCreateRoute,
    ticketDetailsRoute,
    adminRoute,
    profileRoute,
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
