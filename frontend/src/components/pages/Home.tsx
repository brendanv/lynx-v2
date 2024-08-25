import React, { useState } from "react";
import useLinksFeedQuery from "@/hooks/useLinksFeedQuery";
import LinkCard, { LinkCardSkeleton } from "@/components/LinkCard";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import SearchBar, { SearchParams } from "@/components/SearchBar";
import Paginator from "@/components/Paginator";
import { usePageTitle } from "@/hooks/usePageTitle";
import { Toaster } from "@/components/ui/toaster";
import { useSearchParams } from "react-router-dom";
import PageWithHeader from "./PageWithHeader";

const Home: React.FC = () => {
  usePageTitle("My Feed");
  const [urlParams, setUrlParams] = useSearchParams();
  const [page, setPage] = useState(1);
  const [searchParams, setSearchParams] = useState<
    Omit<SearchParams, "searchText">
  >({
    readState: "all",
    sortBy: "added_to_library",
  });
  const searchText = urlParams.get("s") || "";
  const tagId = urlParams.get("t") || undefined;

  const {
    loading: feedLoading,
    result: queryResult,
    error: feedError,
    refetch,
  } = useLinksFeedQuery({
    page,
    ...searchParams,
    searchText,
    tagId,
  });

  const handleSearchParamsChange = (newSearchParams: SearchParams) => {
    setSearchParams(newSearchParams);
    const newParams: {[key: string]: string} = {
      s: newSearchParams.searchText,
    }
    if (newSearchParams.tagId) {
      newParams['t'] = newSearchParams.tagId;
    }
    setUrlParams(newParams);
    setPage(1);
  };

  const renderContent = () => {
    if (feedLoading) {
      return Array.from({ length: 5 }, (_, index) => (
        <LinkCardSkeleton key={`skeleton-${index}`} />
      ));
    }

    if (queryResult && queryResult.items.length > 0) {
      return (
        <>
          {queryResult.items.map((item) => (
            <LinkCard key={item.id} link={item} onUpdate={refetch} />
          ))}
          {queryResult.totalPages > 1 && (
            <Paginator
              currentPage={page}
              totalPages={queryResult.totalPages}
              onPageChange={setPage}
            />
          )}
        </>
      );
    }

    if (feedError || !queryResult || queryResult.items.length === 0) {
      return (
        <Alert variant={feedError ? "destructive" : "default"}>
          <AlertTitle>{feedError ? "Error" : "No links found"}</AlertTitle>
          <AlertDescription>
            {feedError
              ? feedError.message
              : "Try adjusting your filters or add some new links to your feed."}
          </AlertDescription>
        </Alert>
      );
    }

    return null;
  };

  return (
    <PageWithHeader>
      <SearchBar
        searchParams={{ ...searchParams, searchText, tagId }}
        onSearchParamsChange={handleSearchParamsChange}
      />
      <div className="mt-6">{renderContent()}</div>
      <Toaster />
    </PageWithHeader>
  );
};

export default Home;
