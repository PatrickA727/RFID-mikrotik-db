import axios from 'axios';
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from "@tanstack/react-table"
import { Button, ButtonGroup, Text } from "@chakra-ui/react";
import { useState } from "react";

interface ItemSold {
    item_sn: number,
    datetime_sold: Date,
    invoice: string,
    payment_method: string,
    payment_status: string,
}

const ItemSoldTable = () => {
    const [globalFilter, setGlobalFilter] = useState("")
    const [page, setPage] = useState(1)
    const [search, setSearch] = useState("") 

    const limit: number = 10
    let offset: number = 0

    if (page > 1) {
        offset = (page - 1) * limit
    }

    const {data, error, isError, isLoading} = useQuery<{ sold_items: ItemSold[], sold_items_count: number }>({
        queryKey: ["soldItem", offset, search],
        queryFn: async(): Promise<{ sold_items: ItemSold[], sold_items_count: number }> => {
            const { data } = await axios.get<{ sold_items: ItemSold[], sold_items_count: number }>(`api/item/get-sold-items?limit=${limit}&offset=${offset}&search=${search}`);
            return data
        },
        placeholderData: keepPreviousData,
    });

    const totalRecords: number = data?.sold_items_count ?? 0;
    const canPrevPage: boolean = page > 1;
    const canNextPage: boolean = totalRecords > limit * page;
    let totalPages: number = Math.ceil(totalRecords / limit);
    if (totalPages < 1) {
        totalPages = 1
    }

    const columnHelper = createColumnHelper<ItemSold>()

    const columns = [
        columnHelper.accessor("item_sn", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Serial Number
                </span>
            )
        }),

        columnHelper.accessor("datetime_sold", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Date & Time Sold
                </span>
            )
        }),

        columnHelper.accessor("invoice", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Invoice
                </span>
            )
        }),

        columnHelper.accessor("payment_method", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Payment Method
                </span>
            )
        }),

        columnHelper.accessor("payment_status", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Payment Status
                </span>
            )
        }),
    ];

    const sold_items = data?.sold_items ?? [];

    const table = useReactTable({
        data: sold_items,
        columns,
        state: {
            globalFilter
        },
        getCoreRowModel: getCoreRowModel(),
    })

    if (isLoading) {
        return <div>Loading...</div>;
      }
    
    if (isError) {
    return <div>Error: {(error as Error).message}</div>;
    }

    return (
        <div className="flex flex-col min-h-screen max-w-6xl mx-auto py-2 px-4 sm:px-6 lg:px-5">

            <div className="mb-4 relative">
                <input value={search ?? ""} onChange={(e) => setSearch(e.target.value)} placeholder="Search..." className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none"/>
            </div>

            <div className="overflow-x-auto bg-white shadow-md rounded-lg">
                <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                        {
                            table.getHeaderGroups().map((headerGroup) => (
                                <tr key={headerGroup.id}>
                                    {headerGroup.headers.map((header) => (
                                        <th key={header.id} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            <div>
                                                {flexRender(
                                                    header.column.columnDef.header,
                                                    header.getContext()
                                                )}
                                            </div>
                                        </th>
                                    ))}
                                </tr>
                            ))
                        }
                    </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                            {table.getRowModel().rows.map((row) => (
                                <tr key={row.id} className="hover:bg-gray-50">
                                    {row.getVisibleCells().map((cell) => (
                                        <td key={cell.id} className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </td>
                                    ))}
                                </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <Text mb={2} className="mt-2">
                Page {page} of{" "}
                {totalPages}
            </Text>
            <ButtonGroup size="sm" isAttached variant="outline">
                <Button onClick={() => setPage(page - 1)} isDisabled={!canPrevPage} className="mr-1 p-2 rounded-md bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50">
                    {"<"}
                </Button>
                <Button onClick={() => setPage(page + 1)} isDisabled={!canNextPage} className="p-2 rounded-md bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50">
                    {">"}
                </Button>
            </ButtonGroup>
        </div>
    )
}

export default ItemSoldTable