import type { PropsWithChildren, ReactNode } from "react";

type PageLayoutProps = PropsWithChildren<{
  title: string;
  subtitle?: string;
  actions?: ReactNode;
  narrow?: boolean;
}>;

export default function PageLayout({
  title,
  subtitle,
  actions,
  narrow = false,
  children,
}: PageLayoutProps) {
  return (
    <main className="page-shell">
      <section className={`card ${narrow ? "card--narrow" : ""}`}>
        <header className="page-header">
          <div>
            <h1>{title}</h1>
            {subtitle ? <p>{subtitle}</p> : null}
          </div>
          {actions ? <div className="page-header__actions">{actions}</div> : null}
        </header>
        {children}
      </section>
    </main>
  );
}
