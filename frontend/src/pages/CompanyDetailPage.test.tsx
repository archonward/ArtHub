import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useAuth } from "../context/AuthContext";
import { forumApi } from "../services/api/forumApi";
import CompanyDetailPage from "./CompanyDetailPage";

jest.mock("../context/AuthContext", () => ({
  useAuth: jest.fn(),
}));

jest.mock("../services/api/forumApi", () => ({
  forumApi: {
    getCompanyDetails: jest.fn(),
  },
}));

jest.mock(
  "react-router-dom",
  () => ({
    useNavigate: () => jest.fn(),
    useParams: () => ({ id: "7" }),
  }),
  { virtual: true },
);

const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;
const mockedForumApi = forumApi as jest.Mocked<typeof forumApi>;
let consoleErrorSpy: jest.SpyInstance;

const topPosts = [
  {
    id: 1,
    companyId: 7,
    title: "Highest voted",
    body: "Body A",
    createdBy: 2,
    createdAt: "2026-04-06T09:00:00Z",
    voteScore: 5,
    currentUserVote: null,
  },
  {
    id: 2,
    companyId: 7,
    title: "Newest but lower score",
    body: "Body B",
    createdBy: 2,
    createdAt: "2026-04-06T10:00:00Z",
    voteScore: 2,
    currentUserVote: null,
  },
];

const newPosts = [topPosts[1], topPosts[0]];
const firstPageTopPosts = [topPosts[0]];
const secondPageTopPosts = [topPosts[1]];

describe("CompanyDetailPage sorting", () => {
  beforeEach(() => {
    consoleErrorSpy = jest.spyOn(console, "error").mockImplementation((message?: unknown) => {
      if (typeof message === "string" && message.includes("not wrapped in act")) {
        return;
      }
    });

    mockedUseAuth.mockReturnValue({
      currentUser: { id: 2, username: "arthur" },
      isAuthenticated: true,
      isBootstrapping: false,
      authNotice: null,
      login: jest.fn(),
      signup: jest.fn(),
      logout: jest.fn(),
      refreshCurrentUser: jest.fn(),
    });

    mockedForumApi.getCompanyDetails.mockImplementation(
      async (_id: number, sort = "top", page = 1) => ({
        company: {
          id: 7,
          ticker: "AAPL",
          name: "Apple Inc.",
          description: "Discussion",
          createdBy: 2,
          createdAt: "2026-04-06T08:00:00Z",
        },
        posts: sort === "new" ? newPosts : page === 2 ? secondPageTopPosts : firstPageTopPosts,
        pagination: {
          page,
          pageSize: 10,
          totalItems: 2,
          totalPages: sort === "new" ? 1 : 2,
          hasPrev: page > 1,
          hasNext: sort !== "new" && page < 2,
        },
      }),
    );
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
    jest.resetAllMocks();
  });

  it("loads top sort by default and reflects the selected option", async () => {
    await act(async () => {
      render(<CompanyDetailPage />);
    });

    expect(await screen.findByText("Highest voted")).toBeInTheDocument();

    const sortSelect = screen.getByLabelText("Sort");
    expect(sortSelect).toHaveValue("top");
    expect(mockedForumApi.getCompanyDetails).toHaveBeenCalledWith(7, "top", 1, 10);
  });

  it("reloads posts when switching to new sort", async () => {
    await act(async () => {
      render(<CompanyDetailPage />);
    });

    await screen.findByText("Highest voted");

    await userEvent.selectOptions(screen.getByLabelText("Sort"), "new");

    await waitFor(() => {
      expect(mockedForumApi.getCompanyDetails).toHaveBeenLastCalledWith(7, "new", 1, 10);
    });

    const titles = screen.getAllByRole("heading", { level: 3 }).map((node) => node.textContent);
    expect(titles).toEqual(["Newest but lower score", "Highest voted"]);
    expect(screen.getByLabelText("Sort")).toHaveValue("new");
  });

  it("requests the next page and preserves the current sort", async () => {
    await act(async () => {
      render(<CompanyDetailPage />);
    });

    await screen.findByText("Highest voted");

    await userEvent.click(screen.getByRole("button", { name: "Next" }));

    await waitFor(() => {
      expect(mockedForumApi.getCompanyDetails).toHaveBeenLastCalledWith(7, "top", 2, 10);
    });

    expect(screen.getByText("Newest but lower score")).toBeInTheDocument();
    expect(screen.getByText("Page 2 of 2")).toBeInTheDocument();
  });
});
