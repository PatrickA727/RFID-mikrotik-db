import { useCallback, useState } from "react"
import { keepPreviousData, useQuery, useMutation, useQueryClient   } from "@tanstack/react-query";
// import axios from "axios";
import { debounce } from 'lodash';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FaChevronLeft } from 'react-icons/fa';
import { useNavigate } from 'react-router-dom';
import api from "../components/AxiosInstance";
import { toast } from "react-toastify";

interface Item {
  id: number,
  serial_number: string,
  rfid_tag: string,
  type_ref: string
}

interface ItemSold {
  item_tags: string[],
  invoice: string,
  ol_shop: string
}

// const selectedItems = new Set<Item>();

const SellPage = () => {
  const [serialNum, setSerialNum] = useState('')
  const [invoice, setInvoice] = useState('')
  const [ol_shop, setOlShop] = useState('')

  const navigate = useNavigate()

  const [selectedItems, setSelectedItems] = useState<Set<Item>>(new Set());

  const handleAddItem = (item: Item) => {
    const updatedSet = new Set(selectedItems);
    updatedSet.add(item);
    setSelectedItems(updatedSet);
  };

  const handleRemoveItem = (serial_num: string) => {
    const updatedSet = new Set(selectedItems);

  for (const single_item of updatedSet) {
    if (single_item.serial_number === serial_num) {
      updatedSet.delete(single_item); 
      break; 
    }
  }

  // Update the state with the modified Set
  setSelectedItems(updatedSet);
  }

  const queryClient = useQueryClient()

  const handleInvoiceChange = (e) => {
    setInvoice(e.target.value)
  }

  const handleOlshopChange = (e) => {
    setOlShop(e.target.value)
  }

  const debouncedSetSearch = useCallback(
    debounce((query) => setSerialNum(query), 25),
    []
  )

  const handleSearchChange = (e) => {
    debouncedSetSearch(e.target.value)
  }

  const sellItemHandler = async (items:Set<Item>, invoice: string, ol_shop: string) => {
    const rfid_array: string[] = Array.from(items).map((items) => items.rfid_tag)
    if (invoice && ol_shop && rfid_array.length > 0) {
      if (window.confirm(`Item's serial number, invoice, and online shop correct?`)) {
        const newSoldItems: ItemSold = {
          item_tags: rfid_array,
          invoice: invoice,
          ol_shop: ol_shop
        }
  
        try {
          await sellItemMutation(newSoldItems)
          console.log("items sold: ", newSoldItems)
          setInvoice('')
          setOlShop('')
          setSelectedItems(new Set())
          toast.success('Item Sold')
        } catch (error) {
          console.log("error selling items: ", error)
          toast.error("error")
        }
      } else {
        console.log("denied")
      }
    } else {
      return
    }
  } 

  const sellItem = async (item_sold:ItemSold) => {
    try {
      const response = await api.post<ItemSold>(
        `api/item/item-sold-bulk`,
        item_sold
      )
      console.log("POST response: ", response);
    } catch (error) {
      console.log("error selling item: ", error)
    };
  }

  const { mutate: sellItemMutation } = useMutation ({
    mutationFn: sellItem, 
    onSuccess: () => {
      queryClient.invalidateQueries({queryKey: ['items']})
    },
    onError: (error) => {
      console.log("error selling: ", error.message)
    }
  })

  const { data, error, isLoading, isError } = useQuery<{ items: Item[] }>({
    queryKey: ['items', serialNum],
    queryFn: async (): Promise<{items: Item[]}> => {
      const { data } = await api.get<{ items: Item[] }>(`/api/item/get-avail-item?search=${serialNum}`)
      return data
    },
    enabled: serialNum.trim() !== '',
    placeholderData: keepPreviousData,
  })

  const items = data?.items
  console.log(items)

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error: {(error as Error).message}</div>;
  }

  return (
    <div className='bg-gray-200'>
      <div className='px-4 pt-2'>
            <button className="px-2 py-1" onClick={() => navigate("/home")}>
                <FaChevronLeft size={25} />
            </button>
        </div>

        <div className="flex min-h-screen flex-col justify-top py-20 px-60">
          <div className="text-4xl">
            Register Sold Item
          </div>

          <div className="mt-2 relative">
                Invoice:
                <input value={invoice ?? ""} onChange={handleInvoiceChange} placeholder="Order Invoice.." className="w-full pl-5 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none mt-0.5"/>
          </div>

          <div className="mt-2 relative">
                Online Shop:
                <input value={ol_shop ?? ""} onChange={handleOlshopChange} placeholder="Online Shop.." className="w-full pl-5 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none mt-0.5"/>
          </div>

          <div className="mt-2 relative">
                Item's SN:
                <input value={serialNum ?? ""} autoFocus onChange={handleSearchChange} placeholder="Serial Number.." className="w-full pl-5 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none mt-0.5"/>
          </div>

          {serialNum.trim() !== '' && items && items.length > 0 && (
              <div className="bg-white rounded-md border border-gray-300 shadow-lg mt-1">
                <ul className="py-1">
                  {items.map((item) => (
                    <li key={item.id}>
                      <button
                        onClick={() => {
                          handleAddItem(item)
                          console.log("SELECTED: ", selectedItems)
                        }} 
                        className="block w-full px-4 py-2 text-sm text-left text-gray-700 hover:bg-gray-100"
                      >
                        {item.serial_number} - {item.type_ref}
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {selectedItems.size > 0 && (
              <div className="mt-2">
                <ul>
                  {Array.from(selectedItems).map((item) => (
                    <li key={item.id}>
                      <span className="flex justify-between items-center">
                        <span>
                          {item.serial_number} - {item.type_ref}
                        </span>
                        <button onClick={() => handleRemoveItem(item.serial_number)}>
                          <FontAwesomeIcon icon={faTrash} />
                        </button>
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            <button
              onClick={() => sellItemHandler(selectedItems, invoice, ol_shop)} 
              className="block w-40 py-1.5 text-lg text-center text-gray-700 hover:bg-gray-100 bg-white mt-2 rounded-md border border-gray-200">
                Sell
            </button>
        </div>
    </div>
  )
}

export default SellPage
